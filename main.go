package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cloudfoundry/bosh-s3cli/client"
	"github.com/cloudfoundry/bosh-s3cli/config"
)

var version string

func main() {
	configPath := flag.String("c", "", "configuration path")
	showVer := flag.Bool("v", false, "version")
	flag.Parse()

	nonFlagArgs := flag.Args()
	if len(nonFlagArgs) < 2 {
		log.Fatalf("Expected at least two arguments got %d\n", len(nonFlagArgs))
	}

	cmd := nonFlagArgs[0]

	if *showVer {
		fmt.Printf("version %s\n", version)
		os.Exit(0)
	}

	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatalln(err)
	}

	s3Config, err := config.NewFromReader(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	var s3Client *s3.Client
	if cmd != "sign" {
		s3Client, err = client.NewAwsS3Client(&s3Config)
	} else {
		s3Client, err = client.NewAwsS3ClientWithoutAcceptEncodingMiddleware(&s3Config)
	}
	if err != nil {
		log.Fatalln(err)
	}

	blobstoreClient := client.New(s3Client, &s3Config)

	switch cmd {
	case "put":
		if len(nonFlagArgs) != 3 {
			log.Fatalf("Put method expected 3 arguments got %d\n", len(nonFlagArgs))
		}
		src, dst := nonFlagArgs[1], nonFlagArgs[2]

		var sourceFile *os.File
		sourceFile, err = os.Open(src)
		if err != nil {
			log.Fatalln(err)
		}

		defer sourceFile.Close() //nolint:errcheck
		err = blobstoreClient.Put(sourceFile, dst)
	case "get":
		if len(nonFlagArgs) != 3 {
			log.Fatalf("Get method expected 3 arguments got %d\n", len(nonFlagArgs))
		}
		src, dst := nonFlagArgs[1], nonFlagArgs[2]

		var dstFile *os.File
		dstFile, err = os.Create(dst)
		if err != nil {
			log.Fatalln(err)
		}

		defer dstFile.Close() //nolint:errcheck
		err = blobstoreClient.Get(src, dstFile)
	case "delete":
		if len(nonFlagArgs) != 2 {
			log.Fatalf("Delete method expected 2 arguments got %d\n", len(nonFlagArgs))
		}

		err = blobstoreClient.Delete(nonFlagArgs[1])
	case "exists":
		if len(nonFlagArgs) != 2 {
			log.Fatalf("Exists method expected 2 arguments got %d\n", len(nonFlagArgs))
		}

		var exists bool
		exists, err = blobstoreClient.Exists(nonFlagArgs[1])

		// If the object exists the exit status is 0, otherwise it is 3
		// We are using `3` since `1` and `2` have special meanings
		if err == nil && !exists {
			os.Exit(3)
		}
	case "sign":
		if len(nonFlagArgs) != 4 {
			log.Fatalf("Sign method expects 3 arguments got %d\n", len(nonFlagArgs)-1)
		}

		objectID, action := nonFlagArgs[1], nonFlagArgs[2]

		if action != "get" && action != "put" {
			log.Fatalf("Action not implemented: %s. Available actions are 'get' and 'put'", action)
		}

		expiration, err := time.ParseDuration(nonFlagArgs[3])
		if err != nil {
			log.Fatalf("Expiration should be in the format of a duration i.e. 1h, 60m, 3600s. Got: %s", nonFlagArgs[3])
		}

		signedURL, err := blobstoreClient.Sign(objectID, action, expiration)

		if err != nil {
			log.Fatalf("Failed to sign request: %s", err)
			os.Exit(1)
		}

		fmt.Println(signedURL)
		os.Exit(0)
	default:
		log.Fatalf("unknown command: '%s'\n", cmd)
	}

	if err != nil {
		log.Fatalf("performing operation %s: %s\n", cmd, err)
	}
}
