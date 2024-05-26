package utils

import (
	"fmt"
	"strings"
)

// PlainToTXTRecordValue Converts a plain string to a TXT record value.
// if the value in a TXT record is longer than 255 bytes, it needs to be split into multiple parts.
// each part needs to be enclosed in double quotes and separated by a space.
//
// https://datatracker.ietf.org/doc/html/rfc4408#section-3.1.3
func PlainToTXTRecordValue(value string) string {
	if len(value) < 255 {
		return value
	}

	if isChunkedTXTRecordValue(value) {
		return value
	}

	record := strings.Builder{}

	for _, chunk := range chunkSlice(value, 255) {
		record.WriteString(fmt.Sprintf("%q ", chunk))
	}

	return record.String()
}

// TXTRecordToPlainValue converts a TXT record value to a plain string.
// It reverses the operation of PlainToTXTRecordValue.
func TXTRecordToPlainValue(value string) string {
	if !isChunkedTXTRecordValue(value) {
		return value
	}

	record := strings.Builder{}

	for _, chunk := range strings.Fields(value) {
		record.WriteString(unescapeString(chunk))
	}

	return record.String()
}

// isChunkedTXTRecordValue checks if the value is a chunked TXT record value. A chunked TXT record value is a string
// that starts with a double quote and ends with a double quote and optionally a space.
func isChunkedTXTRecordValue(value string) bool {
	return strings.HasPrefix(value, `"`) && (strings.HasSuffix(value, `" `) || strings.HasSuffix(value, `"`))
}

func chunkSlice(slice string, chunkSize int) []string {
	var chunks []string

	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

//nolint:gochecknoglobals
var unescapeReplacer = strings.NewReplacer(`"`, ``, `\"`, `"`)

func unescapeString(value string) string {
	return strings.TrimSpace(unescapeReplacer.Replace(value))
}
