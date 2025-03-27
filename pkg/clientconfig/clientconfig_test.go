package clientconfig

import (
	"testing"

	"github.com/foomo/webgrapple/pkg/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	serviceConfigServices, errRead := readConfigBytes(configBytesServices)
	require.NoError(t, errRead)
	assert.Equal(t, vo.ServiceID("hello"), serviceConfigServices[1].ID)
	serviceConfigService, errRead := readConfigBytes(configBytesService)
	require.NoError(t, errRead)
	assert.Equal(t, vo.ServiceID("my-service"), serviceConfigService[0].ID)
}
