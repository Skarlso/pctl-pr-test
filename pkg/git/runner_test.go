package git_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("runner", func() {
	When("there are git activities to perform", func() {
		It("can abstract how those activities are performed", func() {
			runner := &stringRunner{}
			out, err := runner.Run("cmd", "args1", "args2")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal([]byte("cmd: cmd|args: args1,args2")))
		})
	})
})
