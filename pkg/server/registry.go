package server

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/foomo/webgrapple/pkg/log"
	"github.com/foomo/webgrapple/pkg/vo"
)

func (sm ServiceMap) cp() (copy ServiceMap) {
	copy = ServiceMap{}
	for id, s := range sm {
		copy[id] = s
	}
	return
}

type registryState struct {
	services   ServiceMap
	middleware Middleware
}

type registry struct {
	backendURL        *url.URL
	state             *registryState
	logger            log.Logger
	middlewareFactory WebGrappleMiddleWareCreator
}

func newRegistry(l log.Logger, backendURL *url.URL, middlewareFactory WebGrappleMiddleWareCreator) *registry {
	return &registry{
		logger:            l,
		backendURL:        backendURL,
		middlewareFactory: middlewareFactory,
	}
}

func (r *registry) getServicesCopy() ServiceMap {
	copy := ServiceMap{}
	if r.state != nil && r.state.services != nil {
		copy = r.state.services.cp()
	}
	return copy
}

func (r *registry) upsert(services []*vo.Service) (err error) {
	copy := r.getServicesCopy()
	for _, service := range services {
		r.logger.Info(fmt.Sprintf("upserting service %q with backend %q", service.ID, service.Address))
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
		r.logger.Info(fmt.Sprintf("removing service with ID %q", id))
		delete(copy, id)
	}
	return r.update(copy)
}

func (r *registry) update(services ServiceMap) error {
	newMiddleWare, errCreateMiddleWare := r.middlewareFactory(services, r.backendURL)
	if errCreateMiddleWare != nil {
		return errCreateMiddleWare
	}
	newState := &registryState{
		services:   services,
		middleware: newMiddleWare,
	}
	r.state = newState
	return nil
}
