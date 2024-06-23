package main

import (
	"context"
	ptahAgent "github.com/ptah-sh/ptah-agent/internal/app/ptah-agent"
	"log"
)

// TODO: apply this on CI
//
//	go build -o mybinary \
//	  -ldflags "-X main.version=1.0.0" \
//	  main.go
var version string = "dev"

func main() {
	baseUrl := "http://localhost:8000/api/_nodes/v1"
	ptahToken := "LO0ey15lB2jbObcLWsAUxSIHG3PsNGTR1HQbIwUkju"

	agent := ptahAgent.New(version, baseUrl, ptahToken)

	err := agent.Start(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
}
