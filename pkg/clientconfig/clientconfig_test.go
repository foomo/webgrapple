package clientconfig

import (
	"testing"

	"github.com/foomo/webgrapple/pkg/vo"
	"github.com/stretchr/testify/assert"
)

var (
	configBytesService = []byte(`
---
id: my-service
...
`)
	configBytesServices = []byte(`
---
- id: my-service
- id: hello
...
`)
)

func TestReadConfigService(t *testing.T) {
	serviceConfigServices, errRead := readConfig(configBytesServices)
	assert.NoError(t, errRead)
	assert.Equal(t, vo.ServiceID("hello"), serviceConfigServices[1].ID)
	serviceConfigService, errRead := readConfig(configBytesService)
	assert.NoError(t, errRead)
	assert.Equal(t, vo.ServiceID("my-service"), serviceConfigService[0].ID)
	return
}
