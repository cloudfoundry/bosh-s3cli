package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	s3clicmd "s3cli/cmd"
	fakecmd "s3cli/cmd/fakes"
)

var _ = Describe("runner", func() {
	Describe("Run", func() {
		It("run specified command with given arguments", func() {
			fakeFactory := &fakecmd.FakeFactory{CreatedCmd: &fakecmd.FakeCmd{}}
			runner := s3clicmd.NewRunner(fakeFactory)

			err := runner.Run("some-cmd", []string{"param1", "param2"})
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeFactory.CreatedCmdName).To(Equal("some-cmd"))
			Expect(fakeFactory.CreatedCmd.RunArgs).To(Equal([]string{"param1", "param2"}))
		})
	})
})
