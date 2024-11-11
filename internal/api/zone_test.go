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

	assert.JSONEq(t, expectedJSON, string(computedJSON))
}

func TestCreateZoneRequestJson(t *testing.T) {
	req := CreateZoneRequest{Name: "aName", TTL: 60}
	expectedJSON := `{"name":"aName","ttl":60}`

	assertSerializeAndAssertEqual(t, req, expectedJSON)
}

func TestGetZoneResponseJson(t *testing.T) {
	resp := GetZoneResponse{Zone: Zone{ID: "aId", Name: "aName", TTL: 60, NS: []string{"ns1", "ns2"}}}
	expectedJSON := `{"zone":{"id":"aId","name":"aName","ns":["ns1","ns2"],"ttl":60}}`

	assertSerializeAndAssertEqual(t, resp, expectedJSON)
}

func TestGetZoneByNameResponseJson(t *testing.T) {
	resp := GetZonesByNameResponse{[]Zone{{ID: "aId", Name: "aName", TTL: 60, NS: []string{"ns1", "ns2"}}}}
	expectedJSON := `{"zones":[{"id":"aId","name":"aName","ns":["ns1","ns2"],"ttl":60}]}`

	assertSerializeAndAssertEqual(t, resp, expectedJSON)
}

func TestCreateZoneResponseJson(t *testing.T) {
	resp := CreateZoneResponse{Zone: Zone{ID: "aId", Name: "aName", TTL: 60, NS: []string{"ns1", "ns2"}}}
	expectedJSON := `{"zone":{"id":"aId","name":"aName","ns":["ns1","ns2"],"ttl":60}}`

	assertSerializeAndAssertEqual(t, resp, expectedJSON)
}
