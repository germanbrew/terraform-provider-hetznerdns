package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertSerializeAndAssertEqual(t *testing.T, o interface{}, expectedJSON string) {
	computedJSON, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, string(computedJSON), expectedJSON)
}

func TestCreateZoneRequestJson(t *testing.T) {
	aTTL := int64(60)
	req := CreateZoneRequest{Name: "aName", TTL: &aTTL}
	expectedJSON := `{"name":"aName","ttl":60}`

	assertSerializeAndAssertEqual(t, req, expectedJSON)
}

func TestGetZoneResponseJson(t *testing.T) {
	aTTL := int64(60)
	resp := GetZoneResponse{Zone: Zone{ID: "aId", Name: "aName", TTL: &aTTL}}
	expectedJSON := `{"zone":{"id":"aId","name":"aName","ttl":60}}`

	assertSerializeAndAssertEqual(t, resp, expectedJSON)
}

func TestGetZoneByNameResponseJson(t *testing.T) {
	aTTL := int64(60)
	resp := GetZonesByNameResponse{[]Zone{{ID: "aId", Name: "aName", TTL: &aTTL}}}
	expectedJSON := `{"zones":[{"id":"aId","name":"aName","ttl":60}]}`

	assertSerializeAndAssertEqual(t, resp, expectedJSON)
}

func TestCreateZoneResponseJson(t *testing.T) {
	aTTL := int64(60)
	resp := CreateZoneResponse{Zone: Zone{ID: "aId", Name: "aName", TTL: &aTTL}}
	expectedJSON := `{"zone":{"id":"aId","name":"aName","ttl":60}}`

	assertSerializeAndAssertEqual(t, resp, expectedJSON)
}
