package main

import (
	"context"
	ptahAgent "github.com/ptah-sh/ptah-agent/internal/app/ptah-agent"
	"log"
)

var version string = "dev"

func main() {
	baseUrl := "http://localhost:8000/api/_nodes/v1"
	ptahToken := "aSpD6hq28mIUbdfP0lh6HgqGXVNQRv4SLwNCTHLwFh"

	agent := ptahAgent.New(version, baseUrl, ptahToken)

	err := agent.Start(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
}
