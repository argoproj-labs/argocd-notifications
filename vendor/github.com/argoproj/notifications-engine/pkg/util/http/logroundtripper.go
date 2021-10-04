package http

import (
	"net/http"
	"net/http/httputil"

	log "github.com/sirupsen/logrus"
)

func NewLoggingRoundTripper(roundTripper http.RoundTripper, entry *log.Entry) http.RoundTripper {
	return &logRoundTripper{roundTripper: roundTripper, entry: entry}
}

type logRoundTripper struct {
	roundTripper http.RoundTripper
	entry        *log.Entry
}

func (rt *logRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if info, err := httputil.DumpRequest(req, true); err == nil {
		rt.entry.Debugf("Sending request: %s", string(info))
	}
	resp, err := rt.roundTripper.RoundTrip(req)
	if resp != nil {
		if info, err := httputil.DumpResponse(resp, true); err == nil {
			rt.entry.Debugf("Received response: %s", string(info))
		}
	}
	return resp, err
}
