package clientnpm

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

var errWorkspaceNotFound = errors.New("vscode workspace file not found")

func vscodeGetCodeWorkspace(path string) (string, error) {
	iterator, errIterator := os.ReadDir(path)
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

func vscodedebug(logger Logger, path, name string, debugPort int) error {
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
				_, errRunVSCodeStatus := exec.Command("code", "-s").CombinedOutput()
				if errRunVSCodeStatus != nil {
					logger.Info("waiting for vscode to start...")
				} else {
					logger.Info("vscode is up")
					launchedVSCode = true
					break
				}
			}
			if launchedVSCode {
				logger.Info(fmt.Sprintf("launching vscode: %s", debugConfig))
				combinedOut, errRun := exec.Command(
					"open",
					"vscode://fabiospampinato.vscode-debug-launcher/launch?args="+url.PathEscape(
						debugConfig,
					)).CombinedOutput()
				if errRun != nil {
					logger.Error("could not start vscode %q", string(combinedOut))
				} else {
					logger.Info("started vscode session %q", string(combinedOut))
				}
			}
		} else {
			logger.Error(fmt.Sprintf("could not start vscode due to error: %v and output %q", errLaunch, launchOutput))
		}
	}()
	return nil
}
