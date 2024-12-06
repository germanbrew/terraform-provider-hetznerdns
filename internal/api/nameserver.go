package api

import (
	"context"
	"fmt"
	"net"
)

type Nameserver map[string]string

func resolveIP(ctx context.Context, name string) (map[string]string, error) {
	var (
		resolver   = net.Resolver{PreferGo: true}
		ipv4, ipv6 string
	)

	addrs, err := resolver.LookupIPAddr(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error resolving IP address for %s: %w", name, err)
	}

	for _, addr := range addrs {
		if addr.IP.To4() != nil {
			ipv4 = addr.IP.String()
		} else {
			ipv6 = addr.IP.String()
		}
	}

	return map[string]string{
		"ipv4": ipv4,
		"ipv6": ipv6,
	}, nil
}

func generateNameserverData(ctx context.Context, nameservers []string) ([]Nameserver, error) {
	nsData := make([]Nameserver, 0, len(nameservers))

	for _, ns := range nameservers {
		resolvedNS, err := resolveIP(ctx, ns)
		if err != nil {
			return nil, err
		}

		nsData = append(nsData, Nameserver{
			"name": ns,
			"ipv4": resolvedNS["ipv4"],
			"ipv6": resolvedNS["ipv6"],
		})
	}

	return nsData, nil
}

// GetAuthoritativeNameservers returns a list of all Hetzner DNS authoritative name servers.
// Currently, the list is hard-coded because Hetzner DNS does not provide an API to retrieve this information.
func GetAuthoritativeNameservers(ctx context.Context) ([]Nameserver, error) {
	return generateNameserverData(ctx, []string{
		"helium.ns.hetzner.de.",
		"hydrogen.ns.hetzner.com.",
		"oxygen.ns.hetzner.com.",
	})
}

// GetSecondaryNameservers is a list of all Hetzner DNS secondary name servers.
// Currently, the list is hard-coded because Hetzner DNS does not provide an API to retrieve this information.
func GetSecondaryNameservers(ctx context.Context) ([]Nameserver, error) {
	return generateNameserverData(ctx, []string{
		"ns1.first-ns.de.",
		"robotns2.second-ns.de.",
		"robotns3.second-ns.com.",
	})
}

// GetKonsolehNameservers is a list of all Hetzner DNS KonsoleH name servers.
// Currently, the list is hard-coded because Hetzner DNS does not provide an API to retrieve this information.
func GetKonsolehNameservers(ctx context.Context) ([]Nameserver, error) {
	return generateNameserverData(ctx, []string{
		"ns1.your-server.de.",
		"ns.second-ns.com.",
		"ns3.second-ns.de.",
	})
}
