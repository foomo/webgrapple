package clientconfig

import (
	"testing"

	"github.com/foomo/webgrapple/server"
	"github.com/foomo/webgrapple/vo"
	"github.com/stretchr/testify/assert"
)

var (
	configBytesService = []byte(`
---
path: /foo
id: my-service
...
`)
	configBytesServices = []byte(`
---
-
  id: my-service
  path: /foo/bar
-
  id: hello
  mimetypes:
    - application-x/foo-bar
    - application-x/baz
...
`)
	configBytesMultiServices = []byte(`
---
http://127.0.0.1:8443:
  -
    id: my-service
    path: /foo/bar
http://127.0.0.1:8444:
  -
    id: my-service
    path: /foo/bar
  -
    id: hello
    mimetypes:
      - application-x/foo-bar
      - application-x/baz
...
`)
)

func TestReadConfigService(t *testing.T) {
	serviceConfigMultiServices, errRead := readConfig(configBytesMultiServices)
	assert.NoError(t, errRead)
	assert.Equal(t, vo.ServiceID("hello"), serviceConfigMultiServices["http://127.0.0.1:8444"][1].ID)
	serviceConfigServices, errRead := readConfig(configBytesServices)
	assert.NoError(t, errRead)
	assert.Equal(t, vo.ServiceID("hello"), serviceConfigServices[server.DefaultServiceURL][1].ID)
	serviceConfigService, errRead := readConfig(configBytesService)
	assert.NoError(t, errRead)
	assert.Equal(t, vo.ServiceID("my-service"), serviceConfigService[server.DefaultServiceURL][0].ID)
	return
}
