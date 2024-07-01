package main

import (
	"context"
	"fmt"
	ptahAgent "github.com/ptah-sh/ptah-agent/internal/app/ptah-agent"
	"log"
	"os"
	"path"
	"strings"
)

var version string = "dev"

func main() {
	baseUrl := os.Getenv("PTAH_BASE_URL")
	if baseUrl == "" {
		log.Println("PTAH_BASE_URL is not set, using https://app.ptah.sh")

		baseUrl = "https://app.ptah.sh"
	}

	baseUrl = strings.Trim(baseUrl, "/")
	baseUrl = fmt.Sprintf("%s/api/_nodes/v1", baseUrl)

	ptahToken := os.Getenv("PTAH_TOKEN")
	if ptahToken == "" {
		log.Fatalln("PTAH_TOKEN is not set")
	}

	rootDir := os.Getenv("PTAH_ROOT_DIR")
	if rootDir == "" {
		log.Fatalln("PTAH_ROOT_DIR is not set")
	}

	_, err := os.Stat(rootDir)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = os.Stat(path.Join(rootDir, "versions"))
	if err != nil {
		log.Fatalln("versions dir not found:", err)
	}

	agent := ptahAgent.New(version, baseUrl, ptahToken, rootDir)

	err = agent.Start(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
}
