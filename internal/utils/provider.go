package utils

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ConfigureStringAttribute(attr types.String, envVar, defaultValue string) string {
	if !attr.IsNull() {
		return attr.ValueString()
	}

	if v, ok := os.LookupEnv(envVar); ok {
		return v
	}

	return defaultValue
}

func ConfigureInt64Attribute(attr types.Int64, envVar string, defaultValue int64) (int64, error) {
	if !attr.IsNull() {
		return attr.ValueInt64(), nil
	}

	if v, ok := os.LookupEnv(envVar); ok {
		vInt64, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("error parsing %s: %w", envVar, err)
		}

		return vInt64, nil
	}

	return defaultValue, nil
}

func ConfigureBoolAttribute(attr types.Bool, envVar string, defaultValue bool) (bool, error) {
	if !attr.IsNull() {
		return attr.ValueBool(), nil
	}

	if v, ok := os.LookupEnv(envVar); ok {
		vBool, err := strconv.ParseBool(v)
		if err != nil {
			return false, fmt.Errorf("error parsing %s: %w", envVar, err)
		}

		return vBool, nil
	}

	return defaultValue, nil
}
