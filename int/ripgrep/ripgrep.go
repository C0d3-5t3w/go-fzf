package ripgrep

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
)

// RunRipgrep executes the ripgrep command with the given pattern and directories.
func RunRipgrep(rgPath, pattern string, dirs []string) ([]string, error) {
	if pattern == "" {
		return []string{}, nil // No pattern, no results
	}

	args := []string{"--color", "never", "--line-number", "--no-heading", pattern}
	args = append(args, dirs...)

	cmd := exec.Command(rgPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	// Ripgrep exits with 1 if no matches are found, which is not an error for us.
	// It exits with 2 for actual errors.
	if err != nil && cmd.ProcessState.ExitCode() != 1 {
		return nil, err // Return actual errors
	}

	var results []string
	scanner := bufio.NewScanner(strings.NewReader(stdout.String()))
	for scanner.Scan() {
		results = append(results, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
