package monitor

import (
	"os/exec"
	"strings"
)

var dockerCLIAvailableChecker = func() bool {
	_, err := execLookPath("docker")
	return err == nil
}

// IsDockerCLIAvailable checks availability of the docker CLI in PATH.
func IsDockerCLIAvailable() bool {
	return dockerCLIAvailableChecker()
}

// SetDockerCLIAvailableCheckerForTesting overrides the docker CLI check for unit tests and returns a restore function.
func SetDockerCLIAvailableCheckerForTesting(checker func() bool) func() {
	orig := dockerCLIAvailableChecker
	dockerCLIAvailableChecker = checker
	return func() {
		dockerCLIAvailableChecker = orig
	}
}

// ExecOptions specifies options for launching an interactive session in a container.
type ExecOptions struct {
	Mode      string // "auto", "sh", "zsh", "root", "attach", "custom"
	User      string // optional user (e.g., "0" or "root")
	CustomCmd string // optional custom command when Mode == "custom"
}

// GetContainerExecCommandWithOptions creates an exec.Cmd based on specified ExecOptions.
func GetContainerExecCommandWithOptions(containerID string, opts ExecOptions) *exec.Cmd {
	cleanID := strings.TrimSpace(containerID)
	if cleanID == "" {
		return nil
	}

	if opts.Mode == "attach" {
		return exec.Command("docker", "attach", "--sig-proxy=false", cleanID)
	}

	args := []string{"exec", "-it"}
	if opts.User != "" {
		args = append(args, "-u", opts.User)
	}
	args = append(args, cleanID)

	switch opts.Mode {
	case "sh":
		args = append(args, "sh")
	case "zsh":
		args = append(args, "zsh")
	case "root":
		if opts.User == "" {
			// Insert -u 0 before container ID if not already present
			args = []string{"exec", "-it", "-u", "0", cleanID, "sh", "-c", "if command -v bash >/dev/null 2>&1; then exec bash; else exec sh; fi"}
		} else {
			args = append(args, "sh", "-c", "if command -v bash >/dev/null 2>&1; then exec bash; else exec sh; fi")
		}
	case "custom":
		cmdStr := strings.TrimSpace(opts.CustomCmd)
		if cmdStr == "" {
			cmdStr = "sh"
		}
		args = append(args, "sh", "-c", cmdStr)
	case "auto":
		fallthrough
	default:
		args = append(args, "sh", "-c", "if command -v bash >/dev/null 2>&1; then exec bash; else exec sh; fi")
	}

	return exec.Command("docker", args...)
}

// GetContainerExecCommand creates an exec.Cmd for an interactive container shell.
func GetContainerExecCommand(containerID string) *exec.Cmd {
	return GetContainerExecCommandWithOptions(containerID, ExecOptions{Mode: "auto"})
}
