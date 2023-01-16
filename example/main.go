package main

import (
	"net/http"
	"net/url"

	"github.com/foomo/webgrapple/cmd/webgrapple"
	"github.com/foomo/webgrapple/pkg/server"
	"github.com/foomo/webgrapple/pkg/utils"
	"go.uber.org/zap"
)

func main() {
	logger := utils.GetLogger()
	server.MiddlewareCreator = func(services server.ServiceMap, backendURL *url.URL) (server.Middleware, error) {
		logger.Info("new configuration")
		return func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {

			}
		}, nil
	}
	errExecute := webgrapple.Command.Execute()
	if errExecute != nil {
		utils.GetLogger().Error("execution error", zap.Error(errExecute))
	}
}
