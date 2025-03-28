package clientconfig

import (
	"errors"
	"os"

	"github.com/foomo/webgrapple/pkg/vo"
	"gopkg.in/yaml.v3"
)

func ReadConfig(file string) (multiServerConfig vo.ClientConfig, err error) {
	configBytes, errRead := os.ReadFile(file)
	if errRead != nil {
		return nil, errRead
	}
	return readConfigBytes(configBytes)
}

func readConfigBytes(configBytes []byte) (vo.ClientConfig, error) {
	// a list of services
	clientConfig := vo.ClientConfig{}
	if yaml.Unmarshal(configBytes, &clientConfig) == nil {
		return clientConfig, nil
	}

	// just one service
	var service *vo.Service
	if yaml.Unmarshal(configBytes, &service) == nil {
		return vo.ClientConfig{
			service,
		}, nil
	}

	return nil, errors.New("all attempts to read config failed")
}
