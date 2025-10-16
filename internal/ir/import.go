package ir

import "github.com/google/uuid"

type Import struct {
	ContainerId uuid.UUID `json:"container_id"`
	Target      string    `json:"target"`
	Alias       string    `json:"alias"`
	DetailsJson string    `json:"details_json"`
}
