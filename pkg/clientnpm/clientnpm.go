package clientnpm

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/foomo/webgrapple/pkg/log"
	"github.com/pkg/errors"

	"github.com/foomo/webgrapple/pkg/clientconfig"
	"github.com/foomo/webgrapple/pkg/server"
	"github.com/foomo/webgrapple/pkg/utils/freeport"
	"github.com/foomo/webgrapple/pkg/vo"
)

func getConfig(
	l log.Logger,
	workDir string,
	configPath string,
) (config vo.ClientConfig, err error) {
	if configPath == "" {
		configPath = filepath.Join(workDir, "webgrapple.yaml")
	}

	l.Info(fmt.Sprintf("checking for configuration at path %q", configPath))
	// is there a config
	info, errStat := os.Stat(configPath)
	configExists := errStat == nil && !info.IsDir()
	if configExists {
		l.Info("reading configuration")
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
	_ context.Context,
	l log.Logger,
	flagReverseProxyAddress string,
	flagPort int,
	flagDebugServerPort int,
	flagStartVSCode bool,
	flagConfigPath string,
	workDir string,
	npmCmd string, npmArgs ...string,
) error {
	// setup vars
	name := filepath.Base(workDir)
	l.Info(fmt.Sprintf("starting devproxy client for path %q", flagConfigPath, name))
	config, errGetConfig := getConfig(l, workDir, flagConfigPath)
	if errGetConfig != nil {
		return errorWrap(errGetConfig, "failed to get config webgrapple.yaml is missing ?!")
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
		vscodedebug(l, workDir, name, debugPort)
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
	l.Info("time to register the config with the reverse proxy server(s)")
	errAddServices := addServices(flagReverseProxyAddress, config, port)
	if errAddServices != nil {
		return fmt.Errorf("could not start the app, is the proxy running at %s?", flagReverseProxyAddress)
	}
	defer removeServices(l, flagReverseProxyAddress, config)

	// prepare npm command
	cmd := exec.Command(npmCmd, npmArgs...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), additionalEnvVars...)
	stdOutPipe, errStdOut := cmd.StdoutPipe()
	if errStdOut != nil {
		return errorWrap(errStdOut, "could not open std out")
	}
	stdErrPipe, errStdErr := cmd.StderrPipe()
	if errStdErr != nil {
		return errorWrap(errStdErr, "could not open std err")
	}

	l.Info(fmt.Sprintf("starting npm command '%s %s with env: %s'", npmCmd, strings.Join(npmArgs, " "), strings.Join(additionalEnvVars, "")))
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
			l.Error(fmt.Sprintf("could not copy std out: %v", err))
		}
	}()
	go func() {
		if _, err := io.Copy(os.Stderr, stdErrPipe); err != nil && err.(*os.PathError).Err != os.ErrClosed {
			l.Error(fmt.Sprintf("could not copy std err: %v", err))
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	select {
	case err := <-chanCmdWaitErr:
		if err != nil {
			return errorWrap(err, "command execution failed")
		}
		l.Info("command complete")
	case sig := <-signalChan:
		if err := cmd.Process.Kill(); err != nil {
			return errorWrap(err, "killing child process")
		}
		l.Info(fmt.Sprintf("received signal (%s) interrupt, shutting down gracefully", sig.String()))
	}

	defer l.Info("terminating")
	return nil
}

func removeServices(l log.Logger, address string, config vo.ClientConfig) {
	client := server.NewServiceGoTSRPCClient(string(address), server.DefaultEndPoint)
	serviceIDs := []vo.ServiceID{}
	for _, s := range config {
		serviceIDs = append(serviceIDs, s.ID)
	}
	errRemove, errClient := client.Remove(serviceIDs)
	if errClient != nil {
		l.Error(fmt.Sprintf("could not remove services, got a client error: %v", errClient))
	}
	if errRemove != nil {
		l.Error(fmt.Sprintf("could not remove services due to error: %v", errRemove))
	}
}

func addServices(address string, config vo.ClientConfig, port int) error {
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
