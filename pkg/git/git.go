package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	gitCmd = "git"
)

// Git defines high level abilities for Git related operations.
type Git interface {
	// Add staged changes.
	Add() error
	// Commit changes.
	Commit() error
	// CreateBranch create a branch if it's needed.
	CreateBranch() error
	// CreateRepository bootstraps a plain repository at a given location.
	CreateRepository() error
	// IsRepository returns whether a location is a git repository or not.
	IsRepository() error
	// HasChanges returns whether a location has uncommitted changes or not.
	HasChanges() (bool, error)
	// Push will push to a remote.
	Push() error
}

// CLIGitConfig defines configuration options for CLIGit.
type CLIGitConfig struct {
	Location string
	Branch   string
	Remote   string
	Base     string
}

// CLIGit is a new command line based Git.
type CLIGit struct {
	CLIGitConfig
	Runner Runner
}

// NewCLIGit creates a new command line based Git.
func NewCLIGit(cfg CLIGitConfig, runner Runner) *CLIGit {
	// add .git after location because that's necessary
	cfg.Location = filepath.Join(cfg.Location, ".git")
	return &CLIGit{
		CLIGitConfig: cfg,
		Runner:       runner,
	}
}

// Make sure CLIGit implements all the required methods.
var _ Git = &CLIGit{}

// Commit all changes.
func (g *CLIGit) Commit() error {
	ok, err := g.HasChanges()
	if err != nil {
		return fmt.Errorf("failed to detect if repository has changes: %w", err)
	}
	if !ok {
		// nothing to do.
		return nil
	}
	if err := g.runCmd(gitCmd, "--git-dir", g.Location, "commit", "-am", "Push changes to remote"); err != nil {
		return fmt.Errorf("failed to run commit: %w", err)
	}
	return nil
}

// CreateBranch creates a branch if it differs from the base.
func (g *CLIGit) CreateBranch() error {
	if g.Base == g.Branch {
		return nil
	}
	fmt.Println("creating new branch")
	if err := g.runCmd(gitCmd, "--git-dir", g.Location, "checkout", "-b", g.Branch); err != nil {
		return fmt.Errorf("failed to create new branch %s: %w", g.Branch, err)
	}
	return nil
}

// CreateRepository bootstraps a plain repository at a given location.
func (g *CLIGit) CreateRepository() error {
	return errors.New("implement me")
}

// IsRepository returns whether a location is a git repository or not.
func (g *CLIGit) IsRepository() error {
	// Note that this is redundant in case of CLI git, because the git command line utility
	// already checks if the given location is a repository or not. Never the less we do this
	// for posterity.
	if _, err := os.Stat(g.Location); os.IsNotExist(err) {
		return err
	}
	return nil
}

// HasChanges returns whether a location has uncommitted changes or not.
func (g *CLIGit) HasChanges() (bool, error) {
	out, err := g.Runner.Run(gitCmd, "--git-dir", g.Location, "status", "-s")
	if err != nil {
		return false, fmt.Errorf("failed to check if there are changes: %w", err)
	}
	return string(out) != "", nil
}

// Push will push to a remote.
func (g *CLIGit) Push() error {
	fmt.Println("pushing to remote")
	if err := g.runCmd(gitCmd, "--git-dir", g.Location, "push", g.Remote, g.Branch); err != nil {
		return fmt.Errorf("failed to push changes to remote %s with branch %s: %w", g.Remote, g.Branch, err)
	}
	return nil
}

// Add will add all unstaged changes.
func (g *CLIGit) Add() error {
	fmt.Println("adding unstaged changes")
	if err := g.runCmd(gitCmd, "--git-dir", g.Location, "add", "."); err != nil {
		return fmt.Errorf("failed to run add: %w", err)
	}
	return nil
}

// runCmd is a convenient wrapper around running commands with error output when the output is not needed but needs to
// be logged.
func (g *CLIGit) runCmd(c string, args ...string) error {
	out, err := g.Runner.Run(c, args...)
	if err != nil {
		fmt.Printf("failed to run cmd %s with output: %s\n", c, string(out))
	}
	return err
}
