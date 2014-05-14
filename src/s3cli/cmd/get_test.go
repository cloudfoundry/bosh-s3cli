package cmd_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeclient "s3cli/client/fakes"
	s3clicmd "s3cli/cmd"
)

var _ = Describe("getCmd", func() {
	var (
		client *fakeclient.FakeClient
		cmd    s3clicmd.Cmd
	)

	BeforeEach(func() {
		var err error

		client = &fakeclient.FakeClient{}
		factory := s3clicmd.NewFactory(client)

		cmd, err = factory.Create("get")
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Run", func() {
		Context("with enough arguments", func() {
			It("downloads blob", func() {
				fixtureCatPath := "../../fixtures/cat.jpg"
				tmpCatPath := "../../../tmp/cat.jpg"
				fixtureCatFile, err := os.Open(fixtureCatPath)
				Expect(err).ToNot(HaveOccurred())

				client.GetReaderReadCloser = fixtureCatFile

				err = cmd.Run([]string{"my-cat.jpg", tmpCatPath})
				Expect(err).ToNot(HaveOccurred())

				defer os.RemoveAll(tmpCatPath)

				Expect(client.GetReaderPath).To(Equal("my-cat.jpg"))

				tmpCatFile, err := os.Open(tmpCatPath)
				Expect(err).ToNot(HaveOccurred())

				tmpCatStats, err := tmpCatFile.Stat()
				Expect(err).ToNot(HaveOccurred())
				Expect(tmpCatStats.Size()).To(Equal(int64(1718186)))
			})
		})

		Context("with not enough arguments", func() {
			It("returns error", func() {
				err := cmd.Run([]string{"my-cat.jpg"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(
					"Not enough arguments, expected remote path and destination path"))
			})
		})
	})
})
