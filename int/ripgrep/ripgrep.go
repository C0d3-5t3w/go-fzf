package ripgrep

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
)

type Ripgrep struct {
	Pattern string
	Dirs    []string
}

func FindRipgrep() (string, error) {
	rgPath, err := exec.LookPath("rg")
	if err != nil {
		return "", err 
	}
	return rgPath, nil
}

func RunRipgrep(rgPath, pattern string, dirs []string) ([]string, error) {
	if pattern == "" {
		return []string{}, nil 
	}

	args := []string{"--color", "never", "--line-number", "--no-heading", pattern}
	args = append(args, dirs...)

	cmd := exec.Command(rgPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil && cmd.ProcessState.ExitCode() != 1 {
		return nil, err 
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
