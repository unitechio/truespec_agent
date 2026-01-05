package network

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/net"
)

func NetworkInfo(ctx context.Context, CollectMAC bool) (interface{}, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var networks []map[string]interface{}
	for _, iface := range interfaces {
		netInfo := map[string]interface{}{
			"name":  iface.Name,
			"mtu":   iface.MTU,
			"flags": iface.Flags,
		}

		// Add IP addresses
		var addrs []string
		for _, addr := range iface.Addrs {
			addrs = append(addrs, addr.Addr)
		}
		netInfo["addresses"] = addrs

		// Only include MAC address if policy allows
		if CollectMAC {
			netInfo["mac"] = iface.HardwareAddr
		}

		networks = append(networks, netInfo)
	}

	return map[string]interface{}{
		"interfaces": networks,
	}, nil
}
