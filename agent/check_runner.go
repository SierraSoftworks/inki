package agent

import (
	"os/exec"
	"fmt"
)

func RunChecks() error {
    checks := GetConfig().Checks

    for _, check := range checks {
        c := exec.Command("bash", "-c", check)
        if err := c.Run(); err != nil {
            if exiterr, ok := err.(*exec.ExitError); ok {
                // The program has exited with an exit code != 0

                // This works on both Unix and Windows. Although package
                // syscall is generally platform dependent, WaitStatus is
                // defined for both Unix and Windows and in both cases has
                // an ExitStatus() method with the same signature.
                if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
                    return fmt.Errorf("check '%s' failed with exit code '%d'", check, status.ExitStatus())
                }
            }

            return fmt.Errorf("check '%s' failed ", check)
        }
    }

    return nil
}