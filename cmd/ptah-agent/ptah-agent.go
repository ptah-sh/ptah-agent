package main

import (
	"context"
	"fmt"
	ptahAgent "github.com/ptah-sh/ptah-agent/internal/app/ptah-agent"
	"log"
	"os"
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

	agent := ptahAgent.New(version, baseUrl, ptahToken)

	err := agent.Start(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
}
