package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"s3cli/client"
)

var version string

func main() {
	configPath := flag.String("c", "", "configuration path")
	showVer := flag.Bool("v", false, "version")
	flag.Parse()

	if *showVer {
		fmt.Printf("version %s\n", version)
		os.Exit(0)
	}

	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatalln(err)
	}

	blobstoreClient, err := client.New(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	nonFlagArgs := flag.Args()
	if len(nonFlagArgs) < 2 {
		log.Fatalf("Expected at least two arguments got %d\n", len(nonFlagArgs))
	}

	cmd := nonFlagArgs[0]

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

		defer sourceFile.Close()
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

		defer dstFile.Close()
		err = blobstoreClient.Get(src, dstFile)
	case "delete":
		if len(nonFlagArgs) != 2 {
			log.Fatalf("Delete method expected 2 arguments got %d\n", len(nonFlagArgs))
		}

		err = blobstoreClient.Delete(nonFlagArgs[1])
	case "exist":
		if len(nonFlagArgs) != 2 {
			log.Fatalf("Exist method expected 2 arguments got %d\n", len(nonFlagArgs))
		}

		var exist bool
		exist, err = blobstoreClient.Exist(nonFlagArgs[1])

		// In case object exist return value is 0 and 3 otherwise
		// We are using `3` since `1` and `2` are aleready meaningful used
		if err == nil && !exist {
			os.Exit(3)
		}

	default:
		log.Fatalf("unknown command: '%s'\n", cmd)
	}

	if err != nil {
		log.Fatalf("performing operation %s: %s\n", cmd, err)
	}
}
