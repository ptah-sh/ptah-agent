package networks

import (
	"net"
)

type IP struct {
	Version string `json:"version"`
	IP      string `json:"ip"`
}

type Network struct {
	IfName string `json:"if_name"`
	IPs    []IP   `json:"ips"`
}

func List() ([]Network, error) {
	networks := make([]Network, 0)

	ifs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range ifs {
		isLoopback := i.Flags&net.FlagLoopback == net.FlagLoopback
		isDown := i.Flags&net.FlagUp == 0
		isNotBroadcast := i.Flags&net.FlagBroadcast == 0
		if isLoopback || isDown || isNotBroadcast {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		network := Network{
			IfName: i.Name,
			IPs:    make([]IP, 0),
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip4 := ipNet.IP.To4()
			if ip4 == nil {
				continue
			}

			network.IPs = append(network.IPs, IP{
				Version: "ipv4",
				IP:      ip4.String(),
			})
		}

		if len(network.IPs) > 0 {
			networks = append(networks, network)
		}
	}

	return networks, nil
}
