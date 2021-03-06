package sockclient

import (
	"crypto/tls"
	"errors"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
	"time"
)

const defaultTimeout = 5 * time.Minute

// ErrRedirect ErrRedirect
var ErrRedirect = errors.New("unexpected redirect in response")

// CheckRedirect CheckRedirect
func CheckRedirect(req *http.Request, via []*http.Request) error {
	if via[0].Method == http.MethodGet {
		return http.ErrUseLastResponse
	}
	return ErrRedirect
}

// Get client factory
func Get(u string) *http.Client {
	url, err := url.Parse(u)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"URL": url.RequestURI(),
			"type": "js",
		}).Debugf("Client factory url")
		return nil
	}
	transport := new(http.Transport)
	transport.DisableCompression = true
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	transport.TLSHandshakeTimeout = 30 * time.Second
	transport.IdleConnTimeout = defaultTimeout
	path := url.Path
	transport.Dial = func(_, _ string) (net.Conn, error) {
		return net.DialTimeout("unix", path, defaultTimeout)
	}
	return &http.Client{Transport: transport, CheckRedirect: CheckRedirect, Timeout: defaultTimeout}
}
