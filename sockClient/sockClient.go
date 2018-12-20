package sockclient

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

const defaultTimeout = 32 * time.Second

var ErrRedirect = errors.New("unexpected redirect in response")

func CheckRedirect(req *http.Request, via []*http.Request) error {
	if via[0].Method == http.MethodGet {
		return http.ErrUseLastResponse
	}
	return ErrRedirect
}

// Get lero lero
func Get(u string) *http.Client {
	url, err := url.Parse(u)
	if err != nil {
		log.Debug().Str("URL", url.RequestURI()).Err(err)
		return nil
	}
	transport := new(http.Transport)
	transport.DisableCompression = true
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// transport.DisableKeepAlives = true
	transport.TLSHandshakeTimeout = 5 * time.Second
	// transport.DisableKeepAlives = true
	path := url.Path
	transport.Dial = func(_, _ string) (net.Conn, error) {
		return net.DialTimeout("unix", path, defaultTimeout)
	}
	return &http.Client{Transport: transport, CheckRedirect: CheckRedirect, Timeout: time.Second * 10}
}
