package conditions

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
)

func NewShellScriptConditionFunc() func(context.Context, *corev1alpha2.ShellScriptCondition) (corev1alpha2.ReadinessConditionState, string) {
	return func(ctx context.Context, c *corev1alpha2.ShellScriptCondition) (corev1alpha2.ReadinessConditionState, string) {
		f, err := os.CreateTemp("", "script")
		if err != nil {
			return corev1alpha2.ConditionFailureState, err.Error()
		}

		_, err = f.Write([]byte(c.Script))
		if err != nil {
			return corev1alpha2.ConditionFailureState, err.Error()
		}

		cmd := exec.Command("/bin/bash", f.Name())
		out, _ := cmd.CombinedOutput()
		fmt.Println("##", string(out))
		if cmd.ProcessState.ExitCode() != 0 {
			return corev1alpha2.ConditionFailureState, fmt.Sprintf("Exit code: %d", cmd.ProcessState.ExitCode())
		}

		return corev1alpha2.ConditionSuccessState, ""
	}
}
