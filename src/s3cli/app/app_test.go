package app

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("getConfigPath", func() {
	Context("when -c flag is given", func() {
		It("returns custom path", func() {
			path, args, err := getConfigPath([]string{"-c", "/some/path", "foo", "bar"})
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("/some/path"))
			Expect(args).To(Equal([]string{"foo", "bar"}))
		})
	})

	Context("when -c flag is not given", func() {
		It("returns default path in home directory", func() {
			path, args, err := getConfigPath([]string{"foo", "bar"})
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(ContainSubstring(".s3cli"))
			Expect(args).To(Equal([]string{"foo", "bar"}))
		})
	})
})
