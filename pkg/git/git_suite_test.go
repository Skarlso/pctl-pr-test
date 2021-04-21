package git_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// stringRunner will combine outputs and show what's going to run.
type stringRunner struct {
	output string
	err    error
}

func (s *stringRunner) Run(c string, args ...string) ([]byte, error) {
	out := []byte(fmt.Sprintf("cmd: %s|args: %s", c, strings.Join(args, ",")))
	s.output = string(out)
	return out, s.err
}

func TestGit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Git Suite")
}
