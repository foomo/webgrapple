package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"go.uber.org/zap"
)

const DefaultServiceAddress = "127.0.0.1:8888"
const DefaultServiceURL = "http://127.0.0.1:8888"
const DefaultEndPoint = "/___webgrapple-service"

type srvr struct {
	r                   *registry
	serviceHandler      http.Handler
	defaultProxyHandler http.HandlerFunc
}

func newServer(defaultProxyURL *url.URL, logger *zap.Logger) (*srvr, error) {
	defaultProxy := httputil.NewSingleHostReverseProxy(defaultProxyURL)
	r := newRegistry(logger)
	service := &Service{
		r: r,
	}
	serviceHandler := NewDefaultServiceGoTSRPCProxy(*service, []string{})
	return &srvr{
		r:                   r,
		serviceHandler:      serviceHandler,
		defaultProxyHandler: defaultProxy.ServeHTTP,
	}, nil
}

func (s *srvr) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.r.middleware(s.defaultProxyHandler)(w, r)
}
