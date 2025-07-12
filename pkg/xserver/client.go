package xserver

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	defaultTimeout = 10 * time.Second
)

var (
	ErrInvalidClientOptions = fmt.Errorf("invalid client options: sessionID and deviceKey must not be empty")
)

type Client interface {
	// GetCSRFTokenAsUniqueID retrieves the unique ID for a given VPS ID to be used in extending the VPS expiration.
	GetCSRFTokenAsUniqueID(ctx context.Context, vpsID VPSID) (UniqueID, error)
	// ExtendFreeVPSExpiration extends the expiration of a free VPS.
	ExtendFreeVPSExpiration(ctx context.Context, vpsID VPSID, uniqueID UniqueID) error
}

type ClientOptions struct {
	SessionID string
	DeviceKey string
	Headers   map[string]string
	Logger    *slog.Logger
}

type client struct {
	Logger  *slog.Logger
	Client  *http.Client
	Headers map[string]string
}

var _ Client = (*client)(nil)

func NewClient(options ClientOptions) (Client, error) {
	if options.SessionID == "" || options.DeviceKey == "" {
		return nil, ErrInvalidClientOptions
	}
	if options.Logger == nil {
		options.Logger = slog.Default()
	}

	// Set credentials
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}
	jar.SetCookies(&url.URL{
		Scheme: "https",
		Host:   XServerHost,
	}, []*http.Cookie{
		newCookie("X2SESSID", options.SessionID),
		newCookie("XSERVER_DEVICEKEY", options.DeviceKey),
	})

	// Create HTTP client with the cookie jar
	httpClient := &http.Client{
		Jar: jar,
	}

	return &client{
		Client:  httpClient,
		Logger:  options.Logger,
		Headers: options.Headers,
	}, nil
}

func newCookie(name, value string) *http.Cookie {
	return &http.Cookie{
		Name:   name,
		Value:  value,
		Domain: XServerHost,
		Path:   "/",
		Secure: true,
	}
}

func (c *client) GetCSRFTokenAsUniqueID(ctx context.Context, vpsID VPSID) (UniqueID, error) {
	c.Logger.Info("Retrieving CSRF token for VPS ID", "vpsID", vpsID)
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, FreeVPSExtendURL(vpsID).String(), nil)
	if err != nil {
		return UniqueID(""), fmt.Errorf("failed to create request: %w", err)
	}
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	c.Logger.Debug("Sending request to get CSRF token", "url", req.URL.String())
	resp, err := c.Client.Do(req)
	if err != nil {
		return UniqueID(""), fmt.Errorf("failed to get CSRF token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return UniqueID(""), fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	c.Logger.Debug("Parsing response to find unique ID")
	return findUniqueIdInResponse(resp.Body)
}

func findUniqueIdInResponse(body io.Reader) (UniqueID, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return UniqueID(""), fmt.Errorf("failed to parse response body: %w", err)
	}

	var uniqid string
	doc.Find("input[type=hidden][name=uniqid]").Each(func(i int, s *goquery.Selection) {
		id, exists := s.Attr("value")
		if exists {
			uniqid = id
		}
	})
	if uniqid == "" {
		return UniqueID(""), fmt.Errorf("CSRF token not found in response")
	}

	return UniqueID(uniqid), nil
}

func (c *client) ExtendFreeVPSExpiration(ctx context.Context, vpsID VPSID, uniqueID UniqueID) error {
	c.Logger.Info("Extending free VPS expiration", "vpsID", vpsID, "uniqueID", uniqueID)
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	formData := fmt.Sprintf("uniqid=%s&ethna_csrf=&id_vps=%s", uniqueID, vpsID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, doFreeVPSExtendURL.String(), strings.NewReader(formData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c.Logger.Debug("Sending request to extend VPS expiration", "url", req.URL.String(), "formData", formData)
	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to extend VPS expiration: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}

	c.Logger.Debug("Parsing response to confirm extension")
	body, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(body), "利用期限の更新手続きが完了しました。") {
		c.Logger.Info("VPS expiration extended successfully", "vpsID", vpsID, "uniqueID", uniqueID)
		return nil
	}
	errorMessage, err := findErrorMessageFromResponse(strings.NewReader(string(body)))
	if err != nil {
		c.Logger.Error("Failed to find error message in response", "error", err, "vpsID", vpsID, "uniqueID", uniqueID)
		return fmt.Errorf("VPS renewal failed: %s", err)
	}
	c.Logger.Error("VPS renewal failed", "vpsID", vpsID, "uniqueID", uniqueID, "error_message", errorMessage)
	return fmt.Errorf("VPS renewal failed: %s", errorMessage)
}

func findErrorMessageFromResponse(body io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	var errorMessage []string
	doc.Find("main .contents").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			errorMessage = append(errorMessage, text)
		}
	})

	if len(errorMessage) != 0 {
		return strings.Join(errorMessage, ", "), nil
	}

	doc.Find("main").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			errorMessage = append(errorMessage, text)
		}
	})
	if len(errorMessage) != 0 {
		return strings.Join(errorMessage, ", "), nil
	}

	return "", fmt.Errorf("no error message found in response")
}
