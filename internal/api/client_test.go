package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientCreateZoneSuccess(t *testing.T) {
	t.Parallel()

	var requestBodyReader io.Reader

	responseBody := []byte(`{"zone":{"id":"12345","name":"mydomain.com","ttl":3600}}`)
	config := RequestConfig{responseHTTPStatus: http.StatusOK, requestBodyReader: &requestBodyReader, responseBodyJSON: responseBody}
	client := createTestClient(t, config)

	opts := CreateZoneOpts{Name: "mydomain.com", TTL: 3600}
	zone, err := client.CreateZone(context.Background(), opts)

	require.NoError(t, err)
	assert.Equal(t, Zone{ID: "12345", Name: "mydomain.com", TTL: 3600}, *zone)
	assert.NotNil(t, requestBodyReader, "The request body should not be nil")
	jsonRequestBody, _ := io.ReadAll(requestBodyReader)
	assert.Equal(t, `{"name":"mydomain.com","ttl":3600}`, string(jsonRequestBody))
}

func TestClientCreateZoneInvalidDomain(t *testing.T) {
	t.Parallel()

	//nolint:lll
	responseBody := []byte(`{"zone": {"id":"","name":"","ttl":0,"registrar":"","legacy_dns_host":"","legacy_ns":null,"ns":null,"created":"","verified":"","modified":"","project":"","owner":"","permission":"","zone_type":{"id":"","name":"","description":"","prices":null},"status":"","paused":false,"is_secondary_dns":false,"txt_verification":{"name":"","token":""},"records_count":0},"error":{"message":"422 : invalid TLD","code":422}}`)
	config := RequestConfig{responseHTTPStatus: http.StatusUnprocessableEntity, responseBodyJSON: responseBody}

	client := createTestClient(t, config)
	opts := CreateZoneOpts{Name: "this.is.invalid", TTL: 3600}
	_, err := client.CreateZone(context.Background(), opts)

	require.ErrorContains(t, err, "API returned HTTP 422 Unprocessable Entity error with message: '422 : invalid TLD'")
}

func TestClientCreateZoneInvalidTLD(t *testing.T) {
	t.Parallel()

	var irrelevantConfig RequestConfig
	client := createTestClient(t, irrelevantConfig)
	opts := CreateZoneOpts{Name: "thisisinvalid", TTL: 3600}
	_, err := client.CreateZone(context.Background(), opts)

	require.ErrorContains(t, err, "'thisisinvalid' is not a valid domain")
}

func TestClientUpdateZoneSuccess(t *testing.T) {
	t.Parallel()

	zoneWithUpdates := Zone{ID: "12345678", Name: "zone1.online", TTL: 3600, NS: []string{"ns1.zone1.online", "ns2.zone1.online"}}
	zoneWithUpdatesJSON := `{"id":"12345678","name":"zone1.online","ns":["ns1.zone1.online","ns2.zone1.online"],"ttl":3600}`

	var requestBodyReader io.Reader

	responseBody := []byte(`{"zone":{"id":"12345678","name":"zone1.online","ns":["ns1.zone1.online","ns2.zone1.online"],"ttl":3600}}`)
	config := RequestConfig{responseHTTPStatus: http.StatusOK, requestBodyReader: &requestBodyReader, responseBodyJSON: responseBody}
	client := createTestClient(t, config)

	updatedZone, err := client.UpdateZone(context.Background(), zoneWithUpdates)

	require.NoError(t, err)
	assert.Equal(t, zoneWithUpdates, *updatedZone)
	assert.NotNil(t, requestBodyReader, "The request body should not be nil")
	jsonRequestBody, _ := io.ReadAll(requestBodyReader)
	assert.Equal(t, zoneWithUpdatesJSON, string(jsonRequestBody))
}

func TestClientGetZone(t *testing.T) {
	t.Parallel()

	responseBody := []byte(`{"zone":{"id":"12345678","name":"zone1.online","ttl":3600}}`)
	config := RequestConfig{responseHTTPStatus: http.StatusOK, responseBodyJSON: responseBody}
	client := createTestClient(t, config)

	zone, err := client.GetZone(context.Background(), "12345678")

	require.NoError(t, err)
	assert.Equal(t, Zone{ID: "12345678", Name: "zone1.online", TTL: 3600}, *zone)
}

func TestClientGetZoneReturnNilIfNotFound(t *testing.T) {
	t.Parallel()

	config := RequestConfig{responseHTTPStatus: http.StatusNotFound}
	client := createTestClient(t, config)

	zone, err := client.GetZone(context.Background(), "12345678")

	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, zone)
}

func TestClientGetZoneByName(t *testing.T) {
	t.Parallel()

	responseBody := []byte(`{"zones":[{"id":"12345678","name":"zone1.online","ttl":3600}]}`)
	config := RequestConfig{responseHTTPStatus: http.StatusOK, responseBodyJSON: responseBody}
	client := createTestClient(t, config)

	zone, err := client.GetZoneByName(context.Background(), "zone1.online")

	require.NoError(t, err)
	assert.Equal(t, Zone{ID: "12345678", Name: "zone1.online", TTL: 3600}, *zone)
}

func TestClientGetZoneByNameReturnNilIfnotFound(t *testing.T) {
	t.Parallel()

	config := RequestConfig{responseHTTPStatus: http.StatusNotFound}
	client := createTestClient(t, config)

	zone, err := client.GetZoneByName(context.Background(), "zone1.online")

	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, zone)
}

func TestClientDeleteZone(t *testing.T) {
	t.Parallel()

	config := RequestConfig{responseHTTPStatus: http.StatusOK}
	client := createTestClient(t, config)

	err := client.DeleteZone(context.Background(), "irrelevant")

	require.NoError(t, err)
}

func TestClientGetRecord(t *testing.T) {
	t.Parallel()

	aTTL := int64(3600)
	responseBody := []byte(`{"record":{"zone_id":"wwwlsksjjenm","id":"12345678","name":"zone1.online","ttl":3600,"type":"A","value":"192.168.1.1"}}`)
	config := RequestConfig{responseHTTPStatus: http.StatusOK, responseBodyJSON: responseBody}
	client := createTestClient(t, config)

	record, err := client.GetRecord(context.Background(), "12345678")

	require.NoError(t, err)
	assert.Equal(t, Record{ZoneID: "wwwlsksjjenm", ID: "12345678", Name: "zone1.online", TTL: &aTTL, Type: "A", Value: "192.168.1.1"}, *record)
}

func TestClientGetRecordWithUndefinedTTL(t *testing.T) {
	t.Parallel()

	responseBody := []byte(`{"record":{"zone_id":"wwwlsksjjenm","id":"12345678","name":"zone1.online","type":"A","value":"192.168.1.1"}}`)
	config := RequestConfig{responseHTTPStatus: http.StatusOK, responseBodyJSON: responseBody}
	client := createTestClient(t, config)

	record, err := client.GetRecord(context.Background(), "12345678")

	require.NoError(t, err)
	assert.Equal(t, Record{ZoneID: "wwwlsksjjenm", ID: "12345678", Name: "zone1.online", TTL: nil, Type: "A", Value: "192.168.1.1"}, *record)
}

func TestClientGetRecordReturnNilIfNotFound(t *testing.T) {
	t.Parallel()

	config := RequestConfig{responseHTTPStatus: http.StatusNotFound}
	client := createTestClient(t, config)

	record, err := client.GetRecord(context.Background(), "irrelevant")

	require.Error(t, err)
	assert.Nil(t, record)
}

func TestClientCreateRecordSuccess(t *testing.T) {
	t.Parallel()

	var requestBodyReader io.Reader

	responseBody := []byte(`{"record":{"zone_id":"wwwlsksjjenm","id":"12345678","name":"zone1.online","ttl":3600,"type":"A","value":"192.168.1.1"}}`)
	config := RequestConfig{responseHTTPStatus: http.StatusOK, requestBodyReader: &requestBodyReader, responseBodyJSON: responseBody}
	client := createTestClient(t, config)

	aTTL := int64(3600)
	opts := CreateRecordOpts{ZoneID: "wwwlsksjjenm", Name: "zone1.online", TTL: &aTTL, Type: "A", Value: "192.168.1.1"}
	record, err := client.CreateRecord(context.Background(), opts)

	require.NoError(t, err)
	assert.Equal(t, Record{ZoneID: "wwwlsksjjenm", ID: "12345678", Name: "zone1.online", TTL: &aTTL, Type: "A", Value: "192.168.1.1"}, *record)
	assert.NotNil(t, requestBodyReader, "The request body should not be nil")
	jsonRequestBody, _ := io.ReadAll(requestBodyReader)
	assert.Equal(t, `{"zone_id":"wwwlsksjjenm","type":"A","name":"zone1.online","value":"192.168.1.1","ttl":3600}`, string(jsonRequestBody))
}

func TestClientRecordZone(t *testing.T) {
	t.Parallel()

	config := RequestConfig{responseHTTPStatus: http.StatusOK}
	client := createTestClient(t, config)

	err := client.DeleteRecord(context.Background(), "irrelevant")

	require.NoError(t, err)
}

func TestClientUpdateRecordSuccess(t *testing.T) {
	t.Parallel()

	aTTL := int64(3600)
	recordWithUpdates := Record{ZoneID: "wwwlsksjjenm", ID: "12345678", Name: "zone2.online", TTL: &aTTL, Type: "A", Value: "192.168.1.1"}
	recordWithUpdatesJSON := `{"zone_id":"wwwlsksjjenm","id":"12345678","type":"A","name":"zone2.online","value":"192.168.1.1","ttl":3600}`

	var requestBodyReader io.Reader

	responseBody := []byte(`{"record":{"zone_id":"wwwlsksjjenm","id":"12345678","type":"A","name":"zone2.online","value":"192.168.1.1","ttl":3600}}`)
	config := RequestConfig{responseHTTPStatus: http.StatusOK, requestBodyReader: &requestBodyReader, responseBodyJSON: responseBody}
	client := createTestClient(t, config)

	updatedRecord, err := client.UpdateRecord(context.Background(), recordWithUpdates)

	require.NoError(t, err)
	assert.Equal(t, recordWithUpdates, *updatedRecord)
	assert.NotNil(t, requestBodyReader, "The request body should not be nil")
	jsonRequestBody, _ := io.ReadAll(requestBodyReader)
	assert.Equal(t, recordWithUpdatesJSON, string(jsonRequestBody))
}

func TestClientHandleUnauthorizedRequest(t *testing.T) {
	t.Parallel()

	responseBody := []byte(`{"message":"Invalid API key"}`)
	config := RequestConfig{responseHTTPStatus: http.StatusUnauthorized, responseBodyJSON: responseBody}
	client := createTestClient(t, config)

	opts := CreateZoneOpts{Name: "mydomain.com", TTL: 3600}
	_, err := client.CreateZone(context.Background(), opts)

	require.ErrorContains(t, err, "'Invalid API key'", "Error message didn't contain error message from API.")
}

type RequestConfig struct {
	responseHTTPStatus int
	responseBodyJSON   []byte
	requestBodyReader  *io.Reader
}

func createTestClient(t testing.TB, config RequestConfig) *Client {
	t.Helper()

	client, err := New("http://localhost/", "irrelevant", TestClient{config: config})
	require.NoError(t, err)

	return client
}

type TestClient struct {
	config RequestConfig
}

// See https://golang.org/pkg/net/http/#RoundTripper
func (f TestClient) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil && f.config.requestBodyReader != nil {
		*f.config.requestBodyReader = req.Body
	}

	var jsonBody io.ReadCloser = nil
	if f.config.responseBodyJSON != nil {
		jsonBody = io.NopCloser(bytes.NewReader(f.config.responseBodyJSON))
	}

	resp := http.Response{StatusCode: f.config.responseHTTPStatus, Body: jsonBody}

	return &resp, nil
}
