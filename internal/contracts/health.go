package contracts

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}
