package clientnpm

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/foomo/webgrapple/clientconfig"
	"github.com/foomo/webgrapple/server"
	"github.com/foomo/webgrapple/utils/freeport"
	"github.com/foomo/webgrapple/vo"
	"go.uber.org/zap"
)

func getConfig(
	logger *zap.Logger,
	path string,
) (config vo.ClientConfig, err error) {
	configPath := filepath.Join(path, "webgrapple.yaml")
	logger.Info("checking for configuration",
		zap.String("config-path", configPath),
	)
	// is there a config
	info, errStat := os.Stat(configPath)
	configExists := errStat == nil && !info.IsDir()
	if configExists {
		logger.Info("reading configuration")
		// read config
		config, errConfig := clientconfig.ReadConfig(configPath)
		if errConfig != nil {
			return nil, errorWrap(errConfig, "could not read config from file")
		}
		return config, nil
	}
	return nil, errors.New("config is missing")
}

func errorWrap(err error, wrap string) error {
	return errors.New(wrap + ": " + err.Error())
}

// Run run the command, use this, if Command is in the way
func Run(
	logger *zap.Logger,
	flagReverseProxyAddress string,
	flagPort int,
	flagDebugServerPort int,
	flagStartVSCode bool,
	path string,
	npmCmd string, npmArgs ...string,
) error {

	// setup vars
	name := filepath.Base(path)
	logger.Info(
		"starting devproxy client for",
		zap.String("path", path),
		zap.String("name", path),
	)
	config, errGetConfig := getConfig(logger, path)
	if errGetConfig != nil {
		return errorWrap(errGetConfig, "failed to ge config")
	}

	// get some ports
	port := flagPort
	if port == 0 {
		ports, errTakePort := freeport.Take(1)
		if errTakePort != nil {
			return errorWrap(errTakePort, "could not find a free port")
		}
		port = ports[0]
	}

	debugPort := 0
	if flagDebugServerPort == 0 && flagStartVSCode {
		debugPorts, errTakeDebugPort := freeport.Take(1)
		if errTakeDebugPort != nil {
			return errorWrap(errTakeDebugPort, "could not find a free debug port")
		}
		debugPort = debugPorts[0]
	} else {
		debugPort = flagDebugServerPort
	}
	if flagStartVSCode {
		vscodedebug(logger, path, name, debugPort)
	}

	// ports have to be set in env
	additionalEnvVars := []string{fmt.Sprint("PORT=", port)}
	if debugPort > 0 {
		additionalEnvVars = append(additionalEnvVars, fmt.Sprint("NODE_DEBUG_PORT=", debugPort))
	}

	// pimp config
	for _, service := range config {
		if service.ID == "" {
			service.ID = vo.ServiceID("npm-service-" + name)
		}
		if service.Address == "" {
			// gotta be me
			service.Address = fmt.Sprint("http://127.0.0.1:", port)
		}
	}

	// tell the server about it
	logger.Info("time to register the config with the reverse proxy server(s)")
	errAddServices := addServices(logger, flagReverseProxyAddress, config, port)
	if errAddServices != nil {
		return errorWrap(errAddServices, "could not upsert services to proxy")
	}
	defer removeServices(logger, flagReverseProxyAddress, config)

	// prepare npm command
	cmd := exec.Command(npmCmd, npmArgs...)
	cmd.Env = append(os.Environ(), additionalEnvVars...)
	stdOutPipe, errStdOut := cmd.StdoutPipe()
	if errStdOut != nil {
		return errorWrap(errStdOut, "could not open std out")
	}
	stdErrPipe, errStdErr := cmd.StderrPipe()
	if errStdErr != nil {
		return errorWrap(errStdErr, "could not open std err")
	}

	logger.Info("starting npm command", zap.String("command", npmCmd), zap.Strings("args", npmArgs), zap.Strings("env", additionalEnvVars))
	chanCmdWaitErr := make(chan error)
	if errStart := cmd.Start(); errStart != nil {
		return errorWrap(errStart, fmt.Sprint("faled to start: ", npmCmd, ", with args ", npmArgs))
	}
	defer func() {
		stdOutPipe.Close()
		stdErrPipe.Close()
	}()

	go func() {
		chanCmdWaitErr <- cmd.Wait()
	}()

	go func() {
		if _, err := io.Copy(os.Stdout, stdOutPipe); err != nil && err.(*os.PathError).Err != os.ErrClosed {
			logger.Error("could not copy std out", zap.Error(err))
		}
	}()
	go func() {
		if _, err := io.Copy(os.Stderr, stdErrPipe); err != nil && err.(*os.PathError).Err != os.ErrClosed {
			logger.Error("could not copy std err", zap.Error(err))
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	select {
	case err := <-chanCmdWaitErr:
		if err != nil {
			return errorWrap(err, "command execution failed")
		}
		logger.Info("command complete")
	case sig := <-signalChan:
		if err := cmd.Process.Kill(); err != nil {
			return errorWrap(err, "killing child process")
		}
		logger.Info("received signal interrupt, shutting down gracefully", zap.String("sig", sig.String()))
	}

	defer logger.Info("terminating")
	return nil
}

func removeServices(logger *zap.Logger, address string, config vo.ClientConfig) {
	client := server.NewServiceGoTSRPCClient(string(address), server.DefaultEndPoint)
	serviceIDs := []vo.ServiceID{}
	for _, s := range config {
		serviceIDs = append(serviceIDs, s.ID)
	}
	errRemove, errClient := client.Remove(serviceIDs)
	if errClient != nil {
		logger.Error("could not remove services, got a client error", zap.Error(errClient))
	}
	if errRemove != nil {
		logger.Error("could not remove services", zap.Error(errRemove))
	}
}

func addServices(logger *zap.Logger, address string, config vo.ClientConfig, port int) error {
	client := server.NewServiceGoTSRPCClient(string(address), server.DefaultEndPoint)
	errUpsert, errClient := client.Upsert(config)
	if errClient != nil {
		return errClient
	}
	if errUpsert != nil {
		return errUpsert
	}
	return nil
}
