package server

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/foomo/webgrapple/vo"
	"go.uber.org/zap"
)

type serviceMap map[vo.ServiceID]*vo.Service

func (sm serviceMap) cp() (copy serviceMap) {
	copy = serviceMap{}
	for id, s := range sm {
		copy[id] = s
	}
	return
}

type orderedServiceList []*vo.Service

func (osl orderedServiceList) Len() int           { return len(osl) }
func (osl orderedServiceList) Swap(i, j int)      { osl[i], osl[j] = osl[j], osl[i] }
func (osl orderedServiceList) Less(i, j int) bool { return len(osl[i].Path) < len(osl[j].Path) }

type proxyMap map[vo.ServiceID]*httputil.ReverseProxy

type registryState struct {
	pathServices orderedServiceList
	mimeProxies  map[string]*httputil.ReverseProxy
	services     serviceMap
	proxies      proxyMap
}

type registry struct {
	state  *registryState
	logger *zap.Logger
}

func newRegistry(logger *zap.Logger) *registry {
	return &registry{
		state:  nil,
		logger: logger,
	}
}

func (r *registry) getServicesCopy() serviceMap {
	copy := serviceMap{}
	if r.state != nil {
		copy = r.state.services.cp()
	}
	return copy
}

func (r *registry) upsert(services []*vo.Service) (err error) {
	copy := r.getServicesCopy()
	for _, service := range services {
		r.logger.Info("upserting service", zap.String("id", string(service.ID)), zap.String("path", service.Path), zap.String("backendAddress", service.BackendAddress))
		copy[service.ID] = service
	}
	return r.update(copy)
}

func (r *registry) remove(ids []vo.ServiceID) (err error) {
	copy := r.getServicesCopy()
	for _, id := range ids {
		_, found := copy[id]
		if !found {
			return errors.New("service not found")
		}
		r.logger.Info("removing service", zap.String("id", string(id)))
		delete(copy, id)
	}
	return r.update(copy)
}

func (r *registry) update(services serviceMap) error {
	pathServices := orderedServiceList{}
	proxies := proxyMap{}
	mimeProxies := map[string]*httputil.ReverseProxy{}
	for _, service := range services {
		backendURL, errParseURL := url.Parse(service.BackendAddress)
		if errParseURL != nil {
			return errParseURL
		}
		proxy := httputil.NewSingleHostReverseProxy(backendURL)
		proxies[service.ID] = proxy
		if service.Path != "" {
			pathServices = append(pathServices, service)
		}
		for _, mimetype := range service.MimeTypes {
			mimeProxies[mimetype] = proxy
		}
	}
	r.state = &registryState{
		proxies:      proxies,
		services:     services,
		pathServices: pathServices,
		mimeProxies:  mimeProxies,
	}
	return nil
}

func (r *registry) middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		r.logger.Info("serving", zap.String("path", req.URL.Path))
		if r.state == nil {
			next(rw, req)
			return
		}

		// allows for painless updates in the back
		state := r.state

		// paths are fast and first
		for _, pathService := range state.pathServices {
			if strings.HasPrefix(req.URL.Path, pathService.Path) {
				// req.URL.Path = strings.TrimPrefix(req.URL.Path, pathService.Path)
				state.proxies[pathService.ID].ServeHTTP(rw, req)
				return
			}
		}

		// mime types anyone ?
		mimetype, errGetMimeTypeForPath := r.getMimeTypeForPath(req.URL.Path)
		if errGetMimeTypeForPath != nil {
			http.Error(rw, "could not resolve mime type", http.StatusInternalServerError)
			return
		}
		mimeProxy, okMimeProxy := state.mimeProxies[mimetype]
		if okMimeProxy {
			mimeProxy.ServeHTTP(rw, req)
			return
		}

		// falling back to default backend
		r.logger.Info("falling back to default")
		next(rw, req)
	}
}

func (r *registry) getMimeTypeForPath(path string) (mimeType string, err error) {
	return "implement-me", nil
}
