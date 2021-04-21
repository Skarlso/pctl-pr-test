package git_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/weaveworks/pctl/pkg/git"
)

var _ = Describe("git", func() {
	When("normal flow operations", func() {
		It("can add changes to a commit", func() {
			runner := &stringRunner{}
			g := git.NewCLIGit(git.CLIGitConfig{
				Location: "location",
				Branch:   "main",
				Remote:   "origin",
			}, runner)
			err := g.Add()
			Expect(err).NotTo(HaveOccurred())
			Expect(runner.output).To(Equal("cmd: git|args: --git-dir,location/.git,add,."))
		})
		It("commit changes", func() {
			runner := &stringRunner{}
			g := git.NewCLIGit(git.CLIGitConfig{
				Location: "location",
				Branch:   "main",
				Remote:   "origin",
			}, runner)
			err := g.Commit()
			Expect(err).NotTo(HaveOccurred())
			Expect(runner.output).To(Equal("cmd: git|args: --git-dir,location/.git,commit,-am,Push changes to remote"))
		})
		It("push changes to a remote", func() {
			runner := &stringRunner{}
			g := git.NewCLIGit(git.CLIGitConfig{
				Location: "location",
				Branch:   "main",
				Remote:   "origin",
			}, runner)
			err := g.Push()
			Expect(err).NotTo(HaveOccurred())
			Expect(runner.output).To(Equal("cmd: git|args: --git-dir,location/.git,push,origin,main"))
		})
		It("detects git repositories", func() {
			runner := &stringRunner{}
			tmp, err := ioutil.TempDir("", "detect_git_repo_01")
			Expect(err).NotTo(HaveOccurred())
			err = os.Mkdir(filepath.Join(tmp, ".git"), os.ModeDir)
			Expect(err).NotTo(HaveOccurred())
			g := git.NewCLIGit(git.CLIGitConfig{
				Location: tmp,
				Branch:   "main",
				Remote:   "origin",
			}, runner)
			err = g.IsRepository()
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns an error when the folder is not a git repository", func() {
			runner := &stringRunner{}
			tmp, err := ioutil.TempDir("", "detect_git_repo_02")
			Expect(err).NotTo(HaveOccurred())
			g := git.NewCLIGit(git.CLIGitConfig{
				Location: tmp,
				Branch:   "main",
				Remote:   "origin",
			}, runner)
			err = g.IsRepository()
			Expect(err).To(HaveOccurred())
		})
		It("detects if there are changes to be committed", func() {
			runner := &stringRunner{}
			g := git.NewCLIGit(git.CLIGitConfig{
				Location: "location",
				Branch:   "main",
				Remote:   "origin",
			}, runner)
			ok, err := g.HasChanges()
			Expect(err).ToNot(HaveOccurred())
			Expect(ok).To(BeTrue())
			Expect(runner.output).To(Equal("cmd: git|args: --git-dir,location/.git,status,-s"))
		})
	})
})
