package influxdb

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
)

// RoundTripperFunc is a function which can be used as a http.RoundTripper
type RoundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip delegates to the callee RoundTripperFunc
func (r RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return r(req) }

func opentracingRoundTripper(r http.RoundTripper) http.RoundTripper {
	return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if span := opentracing.SpanFromContext(req.Context()); span != nil {
			opentracing.GlobalTracer().Inject(
				span.Context(),
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(req.Header))
		}

		return r.RoundTrip(req)
	})
}

func newTransport() *http.Transport {
	return &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout:   time.Second * 20,
		Transport: opentracingRoundTripper(newTransport()),
	}
}

// HTTPClientWithTLSConfig returns an *http.Client with sane timeouts and the provided TLSClientConfig.
func HTTPClientWithTLSConfig(conf *tls.Config) *http.Client {
	return &http.Client{
		Timeout: time.Second * 20,
		Transport: opentracingRoundTripper(&http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
			TLSClientConfig:     conf,
		}),
	}
}
