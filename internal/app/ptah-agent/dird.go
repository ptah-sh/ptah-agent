package ptah_agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	t "github.com/ptah-sh/ptah-agent/internal/pkg/ptah-client"
)

func (te *taskExecutor) updateDird(_ context.Context, req *t.UpdateDirdReq) (*t.UpdateDirdRes, error) {
	var tcpPorts, udpPorts []string
	for _, portSpec := range req.NodePorts {
		parts := strings.Split(portSpec, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid port specification: %s", portSpec)
		}
		protocol, port := strings.ToLower(parts[0]), parts[1]
		switch protocol {
		case "tcp":
			tcpPorts = append(tcpPorts, port)
		case "udp":
			udpPorts = append(udpPorts, port)
		default:
			return nil, fmt.Errorf("unknown protocol: %s", protocol)
		}
	}

	ingressGatewayIPs := strings.Join(req.NodeAddresses, ",")
	if ingressGatewayIPs == "" {
		ingressGatewayIPs = "10.0.0.2"
	}
	services := strings.Join(req.DockerServices, ",")

	var args []string
	if len(tcpPorts) > 0 {
		args = append(args, fmt.Sprintf("--tcp-ports %s", strings.Join(tcpPorts, ",")))
	}
	if len(udpPorts) > 0 {
		args = append(args, fmt.Sprintf("--udp-ports %s", strings.Join(udpPorts, ",")))
	}
	args = append(args, fmt.Sprintf("--ingress-gateway-ips %s", ingressGatewayIPs))
	if services != "" {
		args = append(args, fmt.Sprintf("--services %s", services))
	}

	formattedLine := strings.Join(args, " ")

	paramsFilePath := filepath.Join(te.rootDir, "dird", "params.conf")
	if err := os.WriteFile(paramsFilePath, []byte(formattedLine), 0644); err != nil {
		return nil, fmt.Errorf("failed to write to params.conf: %w", err)
	}

	return &t.UpdateDirdRes{}, nil
}
