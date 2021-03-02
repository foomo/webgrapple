package clientnpm

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	"github.com/foomo/webgrapple/clientconfig"
	"github.com/foomo/webgrapple/server"
	"github.com/foomo/webgrapple/utils/freeport"
	"github.com/foomo/webgrapple/vo"
	"go.uber.org/zap"
)

func Run(
	logger *zap.Logger,
	debugServer, startVSCode bool,
	path string,
	npmCmd string, npmArgs ...string,
) error {

	// setup vars
	name := filepath.Base(path)
	configPath := filepath.Join(path, "webgrapple.yaml")
	logger.Info("starting devproxy client for", zap.String("path", path), zap.String("name", path), zap.String("configPath", configPath))

	// read config
	multiServerConfig, errConfig := clientconfig.ReadConfig(configPath)
	if errConfig != nil {
		return errors.Wrap(errConfig, "could not read config from file")
	}

	// get some ports
	ports, errTakePort := freeport.Take(1)
	if errTakePort != nil {
		return errors.Wrap(errTakePort, "could not find a free port")
	}
	port := ports[0]
	debugPort := 0
	if debugServer {
		debugPorts, errTakeDebugPort := freeport.Take(1)
		if errTakeDebugPort != nil {
			return errors.Wrap(errTakeDebugPort, "could not find a free debug port")
		}
		debugPort = debugPorts[0]
		if debugServer && startVSCode {
			vscodedebug(logger, path, name, debugPort)
		}
	}

	// pimp config
	for _, services := range multiServerConfig {
		for _, service := range services {
			if service.BackendAddress == "" {
				// gotta be me
				service.BackendAddress = fmt.Sprint("http://127.0.0.1:", port)
			}
		}
	}

	// tell the server about it
	logger.Info("time to register the config with the reverse proxy server(s)")
	errAddServices := addServices(logger, multiServerConfig, port)
	if errAddServices != nil {
		return errors.Wrap(errAddServices, "could not upsert services to proxy")
	}
	defer removeServices(logger, multiServerConfig)

	// prepare npm command
	cmd := exec.Command(npmCmd, npmArgs...)
	additionalEnvVars := []string{fmt.Sprint("PORT=", port)}
	if debugServer {
		additionalEnvVars = append(additionalEnvVars, fmt.Sprint("NODE_DEBUG_PORT=", debugPort))
	}
	cmd.Env = append(os.Environ(), additionalEnvVars...)
	stdOutPipe, errStdOut := cmd.StdoutPipe()
	if errStdOut != nil {
		return errors.Wrap(errStdOut, "could not open std out")
	}
	stdErrPipe, errStdErr := cmd.StderrPipe()
	if errStdErr != nil {
		return errors.Wrap(errStdErr, "could not open std err")
	}

	logger.Info("starting npm command", zap.String("command", npmCmd), zap.Strings("args", npmArgs), zap.Strings("env", additionalEnvVars))
	chanCmdWaitErr := make(chan error)
	if errStart := cmd.Start(); errStart != nil {
		return errors.Wrapf(errStart, fmt.Sprint("faled to start: ", npmCmd, ", with args ", npmArgs))
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
			return errors.Wrapf(err, "command execution failed")
		}
		fmt.Println("command complete")
	case sig := <-signalChan:
		if err := cmd.Process.Kill(); err != nil {
			return errors.Wrap(err, "killing child process")
		}
		fmt.Println("received signal interrupt,", sig, "shutting down gracefully")
	}

	defer logger.Info("terminating")
	return nil
}
func removeServices(logger *zap.Logger, config vo.MultiServerClientConfig) {
	for address, services := range config {
		client := server.NewServiceGoTSRPCClient(string(address), server.DefaultEndPoint)
		serviceIDs := []vo.ServiceID{}
		for _, s := range services {
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
}

func addServices(logger *zap.Logger, config vo.MultiServerClientConfig, port int) error {
	for address, services := range config {
		client := server.NewServiceGoTSRPCClient(string(address), server.DefaultEndPoint)
		spew.Dump(services)
		errUpsert, errClient := client.Upsert(services)
		if errClient != nil {
			return errClient
		}
		if errUpsert != nil {
			return errUpsert
		}
	}
	return nil
}
