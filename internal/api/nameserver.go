package api

type Nameserver map[string]string

// GetAuthoritativeNameservers returns a list of all Hetzner DNS authoritative name servers.
// Currently, the list is hard-coded because Hetzner DNS does not provide an API to retrieve this information.
func GetAuthoritativeNameservers() []Nameserver {
	return []Nameserver{
		{
			"name": "helium.ns.hetzner.de.",
			"ipv4": "193.47.99.5",
			"ipv6": "2001:67c:192c::add:5",
		},
		{
			"name": "hydrogen.ns.hetzner.com.",
			"ipv4": "213.133.100.98",
			"ipv6": "2a01:4f8:0:1::add:1098",
		},
		{
			"name": "oxygen.ns.hetzner.com.",
			"ipv4": "88.198.229.192",
			"ipv6": "2a01:4f8:0:1::add:2992",
		},
	}
}

// GetSecondaryNameservers is a list of all Hetzner DNS secondary name servers.
// Currently, the list is hard-coded because Hetzner DNS does not provide an API to retrieve this information.
func GetSecondaryNameservers() []Nameserver {
	return []Nameserver{
		{
			"name": "ns1.first-ns.de.",
			"ipv4": "213.239.242.238",
			"ipv6": "2a01:4f8:0:a101::a:1",
		},
		{
			"name": "robotns2.second-ns.de.",
			"ipv4": "213.133.100.103",
			"ipv6": "2a01:4f8:0:1::5ddc:2",
		},
		{
			"name": "robotns3.second-ns.com.",
			"ipv4": "193.47.99.3",
			"ipv6": "2001:67c:192c::add:a3",
		},
	}
}

// GetKonsolehNameservers is a list of all Hetzner DNS KonsoleH name servers.
// Currently, the list is hard-coded because Hetzner DNS does not provide an API to retrieve this information.
func GetKonsolehNameservers() []Nameserver {
	return []Nameserver{
		{
			"name": "ns1.your-server.de.",
			"ipv4": "213.133.100.102",
			"ipv6": "2a01:4f8:0:1::5ddc:1",
		},
		{
			"name": "ns.second-ns.com.",
			"ipv4": "213.239.204.242",
			"ipv6": "2a01:4f8:0:a101::b:1",
		},
		{
			"name": "ns3.second-ns.de.",
			"ipv4": "193.47.99.4",
			"ipv6": "2001:67c:192c::add:b3",
		},
	}
}
