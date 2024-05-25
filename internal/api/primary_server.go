package api

type PrimaryServer struct {
	ID      string `json:"id"`
	Port    uint16 `json:"port"`
	ZoneID  string `json:"zone_id"`
	Address string `json:"address"`
}

type CreatePrimaryServerRequest struct {
	Port    uint16 `json:"port"`
	ZoneID  string `json:"zone_id"`
	Address string `json:"address"`
}

type PrimaryServersResponse struct {
	PrimaryServers []PrimaryServer `json:"primary_servers"`
}

type PrimaryServerResponse struct {
	PrimaryServer PrimaryServer `json:"primary_server"`
}
