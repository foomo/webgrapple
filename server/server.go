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

func newServer(backendURL *url.URL, logger *zap.Logger) (*srvr, error) {
	defaultProxy := httputil.NewSingleHostReverseProxy(backendURL)
	r := newRegistry(logger, backendURL)
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
	if s.r.state != nil && s.r.state.middleware != nil {
		s.r.state.middleware(s.defaultProxyHandler)(w, r)
	} else {
		s.r.logger.Info("you might want to bring up some services, passing request on to backend")
		http.Error(w, "not available - please register at least one service, so that we can bring up your middleware", http.StatusServiceUnavailable)
	}
}
