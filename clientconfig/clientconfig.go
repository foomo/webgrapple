package clientconfig

import (
	"errors"
	"io/ioutil"

	"github.com/foomo/webgrapple/server"
	"github.com/foomo/webgrapple/vo"
	"gopkg.in/yaml.v3"
)

func ReadConfig(file string) (multiServerConfig vo.MultiServerClientConfig, err error) {
	configBytes, errRead := ioutil.ReadFile(file)
	if errRead != nil {
		return nil, errRead
	}
	return readConfig(configBytes)
}

func readConfig(configBytes []byte) (multiServerConfig vo.MultiServerClientConfig, err error) {

	// complex beast mode
	multiServerConfig = vo.MultiServerClientConfig{}
	if yaml.Unmarshal(configBytes, &multiServerConfig) == nil {
		return multiServerConfig, nil
	}

	// a list of service
	clientConfig := vo.ClientConfig{}
	if yaml.Unmarshal(configBytes, &clientConfig) == nil {
		return vo.MultiServerClientConfig{
			server.DefaultServiceURL: clientConfig,
		}, nil
	}

	// just one service
	service := &vo.Service{}
	if yaml.Unmarshal(configBytes, &service) == nil {
		return vo.MultiServerClientConfig{
			server.DefaultServiceURL: vo.ClientConfig{service},
		}, nil
	}

	return nil, errors.New("all attempts to read config failed")
}
