package cmd_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	amzs3 "gopkg.in/amz.v3/s3"

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

		cmd, err = factory.Create("put")
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Run", func() {
		Context("with enough arguments", func() {
			It("uploads blob", func() {
				err := cmd.Run([]string{"../../fixtures/cat.jpg", "my-cat.jpg"})
				Expect(err).ToNot(HaveOccurred())

				file := client.PutReaderReader.(*os.File)
				Expect(client.PutReaderPath).To(Equal("my-cat.jpg"))
				Expect(file.Name()).To(Equal("../../fixtures/cat.jpg"))
				Expect(client.PutReaderLength).To(Equal(int64(1718186)))
				Expect(client.PutReaderContentType).To(Equal("application/octet-stream"))
				Expect(client.PutReaderPerm).To(Equal(amzs3.BucketOwnerFull))
			})
		})

		Context("with not enough arguments", func() {
			It("returns error", func() {
				err := cmd.Run([]string{"my-cat.jpg"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(
					"Not enough arguments, expected source file and destination path"))
			})
		})
	})
})
