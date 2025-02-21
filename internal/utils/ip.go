package utils

import (
	"errors"
	"fmt"
	"net"
)

var ErrInvalidIPAddress = errors.New("invalid IP address")

// CheckIPAddress checks if the given string is a valid IP address.
func CheckIPAddress(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("%w: %s", ErrInvalidIPAddress, ip)
	}

	return nil
}
