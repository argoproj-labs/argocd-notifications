package http

import (
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/argoproj/argo-cd/util/cert"
)

func NewTransport(rawURL string, insecureSkipVerify bool) *http.Transport {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	if insecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return transport
		}
		serverCertificatePem, err := cert.GetCertificateForConnect(parsedURL.Host)
		if err != nil {
			return transport
		} else if len(serverCertificatePem) > 0 {
			transport.TLSClientConfig = &tls.Config{
				RootCAs: cert.GetCertPoolFromPEMData(serverCertificatePem),
			}
		}
	}
	return transport
}
