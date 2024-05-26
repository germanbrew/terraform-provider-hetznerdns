//nolint:lll
package utils_test

import (
	"strings"
	"testing"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestPlainToTXTRecordValue(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "empty",
			input:  "",
			output: "",
		},
		{
			name:   "small string",
			input:  "test",
			output: "test",
		},
		{
			name:   "small string with quotes",
			input:  `te"st`,
			output: `te"st`,
		},
		{
			name:   "large string",
			input:  strings.Repeat("test", 100),
			output: `"testtesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttes" "ttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttest" `,
		},
		{
			name:   "large string with quotes",
			input:  strings.Repeat(`te"st`, 100),
			output: `"te\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"st" "te\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"st" `,
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.output, utils.PlainToTXTRecordValue(tc.input))
		})
	}
}

func TestTXTRecordToPlainValue(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "empty",
			input:  "",
			output: "",
		},
		{
			name:   "small string",
			input:  "test",
			output: "test",
		},
		{
			name:   "small string with quotes",
			input:  `te"st`,
			output: `te"st`,
		},
		{
			name:   "large string",
			input:  `"testtesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttes" "ttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttest" `,
			output: strings.Repeat("test", 100),
		},
		{
			name:   "large string with quotes",
			input:  `"te\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"st" "te\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"stte\"st" `,
			output: strings.Repeat(`te"st`, 100),
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.output, utils.TXTRecordToPlainValue(tc.input))
		})
	}
}

func TestPlainToTXTRecordToPlainValue(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		value string
	}{
		{
			name:  "empty",
			value: "",
		},
		{
			name:  "small string",
			value: "test",
		},
		{
			name:  "small string with quotes",
			value: `te"st`,
		},
		{
			name:  "large string",
			value: strings.Repeat("test", 100),
		},
		{
			name:  "large string with quotes",
			value: strings.Repeat(`te"st`, 100),
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.value, utils.TXTRecordToPlainValue(utils.TXTRecordToPlainValue(tc.value)))
		})
	}
}
