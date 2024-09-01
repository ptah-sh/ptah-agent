package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	ptahAgent "github.com/ptah-sh/ptah-agent/internal/app/ptah-agent"
	"github.com/ptah-sh/ptah-agent/internal/pkg/networks"
)

var version string = "dev"

func main() {
	// Define subcommands
	listIPsCmd := flag.NewFlagSet("list-ips", flag.ExitOnError)
	execTasksCmd := flag.NewFlagSet("exec-tasks", flag.ExitOnError)

	// Parse the main command
	if len(os.Args) == 1 {
		runAgent()

		return
	}

	switch os.Args[1] {
	case "list-ips":
		listIPsCmd.Parse(os.Args[2:])
		listIPs()
	case "exec-tasks":
		execTasksCmd.Parse(os.Args[2:])
		if execTasksCmd.NArg() < 1 {
			log.Fatal("Please provide a JSON file path for exec-tasks")
		}
		execTasks(execTasksCmd.Arg(0))
	default:
		log.Fatalf("Unknown command: %s", os.Args[1])
	}
}

func listIPs() {
	networks, err := networks.List()
	if err != nil {
		log.Fatalf("Error listing networks: %v", err)
	}

	var ips []string
	for _, network := range networks {
		for _, ip := range network.IPs {
			ips = append(ips, ip.IP)
		}
	}

	fmt.Println(strings.Join(ips, " "))
}

func runAgent() {
	baseUrl := os.Getenv("PTAH_BASE_URL")
	if baseUrl == "" {
		log.Println("PTAH_BASE_URL is not set, using https://ctl.ptah.sh")

		baseUrl = "https://ctl.ptah.sh"
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

	agent, err := ptahAgent.New(version, baseUrl, ptahToken, rootDir)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	err = agent.Start(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
}

func execTasks(jsonFilePath string) {
	baseUrl := os.Getenv("PTAH_BASE_URL")
	if baseUrl == "" {
		log.Println("PTAH_BASE_URL is not set, using https://ctl.ptah.sh")
		baseUrl = "https://ctl.ptah.sh"
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

	agent, err := ptahAgent.New(version, baseUrl, ptahToken, rootDir)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	err = agent.ExecTasks(context.Background(), jsonFilePath)
	if err != nil {
		log.Fatalf("Error executing tasks: %v", err)
	}
}
