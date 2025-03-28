# WebGrapple

> A proxy and a client to take over routes of a remote server with local web servers

## Example

```go
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
	webgrapple.MiddlewareFactory = func(services server.ServiceMap, backendURL *url.URL) (server.Middleware, error) {
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
```

## How to Contribute

Please refer to the [CONTRIBUTING](.github/CONTRIBUTING.md) details and follow the [CODE_OF_CONDUCT](.github/CODE_OF_CONDUCT.md) and [SECURITY](.github/SECURITY.md) guidelines.

## License

Distributed under MIT License, please see license file within the code for more details.

_Made with ♥ [foomo](https://www.foomo.org) by [bestbytes](https://www.bestbytes.com)_

