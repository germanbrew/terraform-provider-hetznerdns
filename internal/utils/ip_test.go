package utils_test

import (
	"testing"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestCheckIPAddress(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		ip      string
		isValid bool
	}{
		{
			name:    "valid IPv4",
			ip:      "9.9.9.9",
			isValid: true,
		},
		{
			name:    "invalid IPv4",
			ip:      "9.9.9.999",
			isValid: false,
		},
		{
			name:    "invalid IPv4 with space",
			ip:      "9.9.9.9 ",
			isValid: false,
		},
		{
			name:    "valid IPv6",
			ip:      "2001:4860:4860::8888",
			isValid: true,
		},
		{
			name:    "invalid IPv6",
			ip:      "2001:4860:4860:::8888",
			isValid: false,
		},
		{
			name:    "invalid IPv6 with space",
			ip:      "2001:4860:4860::8888 ",
			isValid: false,
		},
		{
			name:    "invalid IP",
			ip:      "invalid",
			isValid: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := utils.CheckIPAddress(tc.ip)
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
