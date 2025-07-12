package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const defaultTimeout = 10 * time.Second

const (
	XServerHost         = "secure.xserver.ne.jp"
	FreeVPSExtendPath   = "/xapanel/xvps/server/freevps/extend/index"
	DoFreeVPSExtendPath = "/xapanel/xvps/server/freevps/extend/do"
)

var (
	RegExpUniqueID = regexp.MustCompile(`<input type="hidden" name="uniqid" value="([^"]+)" />`)
)

func mustParseURL(rawURL string) url.URL {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		slog.Error("Error parsing URL", "url", rawURL, "error", err)
	}
	return *parsed
}

type XServer struct {
	freeVPSExtendURL url.URL
	doExtendURL      url.URL
	headers          map[string]string
	client           *http.Client
}

func NewXServerClient(sessionID, deviceKey string, headers map[string]string) (*XServer, error) {
	var err error

	x := &XServer{
		headers: headers,
	}

	x.freeVPSExtendURL = mustParseURL("https://" + XServerHost + FreeVPSExtendPath)
	x.doExtendURL = mustParseURL("https://" + XServerHost + DoFreeVPSExtendPath)

	client := &http.Client{
		Timeout: defaultTimeout,
	}
	client.Jar, err = x.NewCookieJar(sessionID, deviceKey)
	if err != nil {
		return nil, err
	}
	x.client = client

	return x, nil
}

func (x *XServer) NewCookie(name, value string) *http.Cookie {
	return &http.Cookie{
		Name:   name,
		Value:  value,
		Domain: XServerHost,
		Path:   "/",
		Secure: true,
	}
}

func (x *XServer) NewCredentials(sessionID, deviceKey string) []*http.Cookie {
	return []*http.Cookie{
		x.NewCookie("X2SESSID", sessionID),
		x.NewCookie("XSERVER_DEVICEKEY", deviceKey),
	}
}

func (x *XServer) NewCookieJar(sessionID, deviceKey string) (*cookiejar.Jar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	jar.SetCookies(&url.URL{
		Scheme: "https",
		Host:   "secure.xserver.ne.jp",
	}, x.NewCredentials(sessionID, deviceKey))

	return jar, nil
}

func (x *XServer) FreeVPSExtendURL(vpsID string) url.URL {
	extendURL := x.freeVPSExtendURL
	q := extendURL.Query()
	q.Add("id_vps", vpsID)
	extendURL.RawQuery = q.Encode()
	return extendURL
}

func (x *XServer) DoFreeVPSExtendURL(vpsID, uniqueID string) url.URL {
	return x.doExtendURL
}

func (x *XServer) GetUniqueID(ctx context.Context, vpsID string) (string, error) {
	slog.Debug("Getting unique ID for VPS", "vps_id", vpsID)
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	// Generate the request URL
	u := x.FreeVPSExtendURL(vpsID)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	for k, v := range x.headers {
		r.Header.Set(k, v)
	}

	// Request to get the unique ID
	slog.Debug("Making request to get unique ID", "url", u.String(), "headers", x.headers)
	resp, err := x.client.Do(r)
	if err != nil {
		slog.Error("Error making request", "error", err, "url", u.String())
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		slog.Error("Unexpected response status", "status_code", resp.StatusCode, "url", u.String(), "body", string(body))
		return "", fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}
	slog.Debug("Request completed successfully", "status_code", resp.StatusCode, "url", u.String())

	matches := RegExpUniqueID.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		slog.Error("Unique ID not found in response", "url", u.String(), "body", string(body))
		return "", fmt.Errorf("unique ID not found in response")
	}
	uniqueID := matches[1]
	slog.Debug("Unique ID found", "unique_id", uniqueID, "vps_id", vpsID)

	return uniqueID, nil
}

func (x *XServer) DoExtendFreeVPS(ctx context.Context, vpsID, uniqueID string) error {
	slog.Debug("Extending free VPS", "vps_id", vpsID, "unique_id", uniqueID)
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	formData := fmt.Sprintf("uniqid=%s&ethna_csrf=&id_vps=%s", uniqueID, vpsID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, x.doExtendURL.String(), strings.NewReader(formData))
	if err != nil {
		slog.Error("Error creating request", "error", err, "vps_id", vpsID, "unique_id", uniqueID)
		return err
	}
	for k, v := range x.headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	slog.Debug("Making request to extend free VPS", "url", x.doExtendURL.String(), "headers", x.headers)

	resp, err := x.client.Do(req)
	if err != nil {
		slog.Error("Error making request to extend free VPS", "error", err, "url", x.doExtendURL.String(), "vps_id", vpsID, "unique_id", uniqueID)
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		slog.Error("Unexpected response status", "status_code", resp.StatusCode, "url", x.doExtendURL.String(), "body", string(body))
		return fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}
	slog.Debug("Request completed successfully", "status_code", resp.StatusCode, "url", x.doExtendURL.String())

	if strings.Contains(string(body), "利用期限の更新手続きが完了しました。") {
		slog.Info("VPS renewal completed successfully", "vps_id", vpsID, "unique_id", uniqueID)
	} else {
		slog.Error("VPS renewal failed", "vps_id", vpsID, "unique_id", uniqueID, "response_text", formatHTMLForLogging(string(body), resp.StatusCode))
		return fmt.Errorf("VPS renewal failed: %s", formatHTMLForLogging(string(body), resp.StatusCode))
	}

	return nil
}
