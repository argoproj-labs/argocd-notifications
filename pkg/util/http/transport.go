package http

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
)

var certResolver func(serverName string) ([]string, error)

func SetCertResolver(resolver func(serverName string) ([]string, error)) {
	certResolver = resolver
}

func NewTransport(rawURL string, insecureSkipVerify bool) *http.Transport {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	if insecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else if certResolver != nil {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return transport
		}
		serverCertificatePem, err := certResolver(parsedURL.Host)
		if err != nil {
			return transport
		} else if len(serverCertificatePem) > 0 {
			transport.TLSClientConfig = &tls.Config{
				RootCAs: getCertPoolFromPEMData(serverCertificatePem),
			}
		}
	}
	return transport
}

func getCertPoolFromPEMData(pemData []string) *x509.CertPool {
	certPool := x509.NewCertPool()
	for _, pem := range pemData {
		certPool.AppendCertsFromPEM([]byte(pem))
	}
	return certPool
}
