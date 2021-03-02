package clientnpm

import (
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"
)

func vscodedebug(logger *zap.Logger, path, name string, debugPort int) error {
	absPath, errAbsConfigPath := filepath.Abs(path)
	if errAbsConfigPath != nil {
		return errAbsConfigPath
	}
	const tries = 5
	go func() {
		logger.Info("starting vscode")
		launchOutput, errLaunch := exec.Command("code", absPath).CombinedOutput()
		if errLaunch == nil {
			// launchedVSCode := false
			// for i := 0; i < 5; i++ {
			// 	zapAttempt := zap.Int("attempt", i)
			// 	vsCodeStatus, errRunVSCodeStatus := exec.Command("code", "-s").CombinedOutput()
			// 	if errRunVSCodeStatus != nil {
			// 		logger.Info("waiting for vscode to start", zapAttempt)
			// 	} else {
			// 		logger.Info("vscode is up", zap.String("status", string(vsCodeStatus)), zapAttempt)
			// 		launchedVSCode = true
			// 		break
			// 	}
			// }
			// if launchedVSCode {
			launchJSON := `{ "sourceMaps": true, "type": "node", "request": "attach", "name": "webgrappleclient - ` + name + `", "port": ` + fmt.Sprint(debugPort) + `}`
			logger.Info("launching vscode", zap.String("json", launchJSON))
			combinedOut, errRun := exec.Command(
				"open",
				"vscode://fabiospampinato.vscode-debug-launcher/launch?args="+url.PathEscape(
					launchJSON,
				)).CombinedOutput()
			zapOutput := zap.String("output", string(combinedOut))
			if errRun != nil {
				logger.Error("could not start vscode", zapOutput)
			} else {
				logger.Info("started vscode session", zapOutput)
			}
			// }
		} else {
			logger.Error("could not start vscode", zap.Error(errLaunch), zap.String("output", string(launchOutput)))
		}
	}()
	return nil
}
