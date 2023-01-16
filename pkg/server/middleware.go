package server

import (
	http "net/http"
	"net/url"

	"github.com/foomo/webgrapple/pkg/vo"
)

// ServiceMap a map of registered services
type ServiceMap map[vo.ServiceID]*vo.Service

// Middleware your way to handle requests
type Middleware func(next http.HandlerFunc) http.HandlerFunc

// WebGrappleMiddleWareCreator create a project specific middleware, when configs change
type WebGrappleMiddleWareCreator func(services ServiceMap, fallbackServerURL *url.URL) (middleware Middleware, errCreation error)

// MiddlewareCreator set this to register you webgrapple middleware creator
var MiddlewareCreator WebGrappleMiddleWareCreator = nil
