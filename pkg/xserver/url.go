package xserver

import (
	"net/url"
)

const (
	XServerHost         = "secure.xserver.ne.jp"
	FreeVPSExtendPath   = "/xapanel/xvps/server/freevps/extend/index"
	DoFreeVPSExtendPath = "/xapanel/xvps/server/freevps/extend/do"
)

var (
	freeVPSExtendURL   = mustJoinURL("https://", XServerHost, FreeVPSExtendPath)
	doFreeVPSExtendURL = mustJoinURL("https://", XServerHost, DoFreeVPSExtendPath)
)

func mustJoinURL(base string, elem ...string) *url.URL {
	u, err := url.Parse(base)
	if err != nil {
		panic(err)
	}
	return u.JoinPath(elem...)
}

func FreeVPSExtendURL(id VPSID) *url.URL {
	u := freeVPSExtendURL
	q := u.Query()
	q.Set("vpsid", id.String())
	u.RawQuery = q.Encode()
	return u
}
