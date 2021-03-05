package server

import (
	"errors"
	"net/http"

	"github.com/foomo/webgrapple/vo"
	"go.uber.org/zap"
)

// ServiceMap a map of registered services
type ServiceMap map[vo.ServiceID]*vo.Service

// Middleware your way to handle requests
type Middleware func(next http.HandlerFunc) http.HandlerFunc

// WebGrappleMiddleWareCreator create a project specific middleware, when configs change
type WebGrappleMiddleWareCreator func(services ServiceMap) (middleware Middleware, errCreation error)

// MiddlewareCreator set this to register you webgrapple middleware creator
var MiddlewareCreator WebGrappleMiddleWareCreator = nil

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
	state  *registryState
	logger *zap.Logger
}

func newRegistry(logger *zap.Logger) *registry {
	return &registry{
		logger: logger,
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
	newMiddleWare, errCreateMiddleWare := MiddlewareCreator(services)
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
