package clientnpm

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var errWorkspaceNotFound = errors.New("vscode workspace file not found")

func vscodeGetCodeWorkspace(path string) (string, error) {
	iterator, errIterator := ioutil.ReadDir(path)
	if errIterator != nil {
		return "", errIterator
	}
	for _, f := range iterator {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".code-workspace") {
			return filepath.Join(path, f.Name()), nil
		}
	}
	dir := filepath.Dir(path)
	if len(dir) == len(path) {
		return "", errWorkspaceNotFound
	}
	return vscodeGetCodeWorkspace(dir)
}

func vscodeGetTarget(path string) string {
	workspaceFile, err := vscodeGetCodeWorkspace(path)
	if err == nil {
		return workspaceFile
	}
	return path
}

func vscodeDebugConfig(name string, debugPort int) (config string, err error) {
	launchData := map[string]interface{}{
		"sourceMaps": true,
		"type":       "node",
		"request":    "attach",
		"name":       "webgrapple-npm " + name,
		"port":       debugPort,
	}
	launchJSONBytes, errMarshal := json.Marshal(launchData)
	if errMarshal != nil {
		return "", errMarshal
	}
	return string(launchJSONBytes), nil
}

func vscodedebug(logger *zap.Logger, path, name string, debugPort int) error {
	absPath, errAbsConfigPath := filepath.Abs(path)
	if errAbsConfigPath != nil {
		return errAbsConfigPath
	}
	vscodeTarget := vscodeGetTarget(absPath)

	debugConfig, errDebugConfig := vscodeDebugConfig(name, debugPort)
	if errDebugConfig != nil {
		return nil
	}

	const tries = 5

	go func() {
		logger.Info("starting vscode")
		launchOutput, errLaunch := exec.Command("code", vscodeTarget).CombinedOutput()
		if errLaunch == nil {
			launchedVSCode := false
			for i := 0; i < 5; i++ {
				zapAttempt := zap.Int("attempt", i)
				vsCodeStatus, errRunVSCodeStatus := exec.Command("code", "-s").CombinedOutput()
				if errRunVSCodeStatus != nil {
					logger.Info("waiting for vscode to start", zapAttempt)
				} else {
					logger.Info("vscode is up", zap.String("status", string(vsCodeStatus)), zapAttempt)
					launchedVSCode = true
					break
				}
			}
			if launchedVSCode {
				logger.Info("launching vscode", zap.String("json", debugConfig))
				combinedOut, errRun := exec.Command(
					"open",
					"vscode://fabiospampinato.vscode-debug-launcher/launch?args="+url.PathEscape(
						debugConfig,
					)).CombinedOutput()
				zapOutput := zap.String("output", string(combinedOut))
				if errRun != nil {
					logger.Error("could not start vscode", zapOutput)
				} else {
					logger.Info("started vscode session", zapOutput)
				}
			}
		} else {
			logger.Error("could not start vscode", zap.Error(errLaunch), zap.String("output", string(launchOutput)))
		}
	}()
	return nil
}
