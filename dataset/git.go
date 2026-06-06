package dataset

import (
	"os/exec"
	"strings"
)

func generatorGitState() (string, bool) {
	commitOutput, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "unknown", true
	}

	statusOutput, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return strings.TrimSpace(string(commitOutput)), true
	}

	return strings.TrimSpace(string(commitOutput)), len(strings.TrimSpace(string(statusOutput))) > 0
}
