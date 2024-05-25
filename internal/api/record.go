package api

// Record represents a record in a specific Zone
type Record struct {
	ZoneID string `json:"zone_id"`
	ID     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Value  string `json:"value"`
	TTL    *int   `json:"ttl,omitempty"`
}

// HasTTL returns true if a Record has a TTL set and false if TTL is undefined
func (r *Record) HasTTL() bool {
	return r.TTL != nil
}

// CreateRecordRequest represents all data required to create a new record
type CreateRecordRequest struct {
	ZoneID string `json:"zone_id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Value  string `json:"value"`
	TTL    *int   `json:"ttl,omitempty"`
}

// RecordsResponse represents a response from tha API containing a list of records
type RecordsResponse struct {
	Records []Record `json:"records"`
}

// RecordResponse represents a response from the API containing only one record
type RecordResponse struct {
	Record Record `json:"record"`
}
