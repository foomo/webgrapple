package server

import (
	"errors"
	"net/url"

	"github.com/foomo/webgrapple/pkg/vo"
	"go.uber.org/zap"
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
	backendURL *url.URL
	state      *registryState
	logger     *zap.Logger
}

func newRegistry(logger *zap.Logger, backendURL *url.URL) *registry {
	return &registry{
		logger:     logger,
		backendURL: backendURL,
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
		r.logger.Info(
			"upserting service",
			zap.String("id", string(service.ID)),
			zap.String("backendAddress", service.Address),
		)
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

func (r *registry) update(services ServiceMap) error {
	newMiddleWare, errCreateMiddleWare := MiddlewareCreator(services, r.backendURL)
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
