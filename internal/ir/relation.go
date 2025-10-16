package ir

import "github.com/google/uuid"

type Relation struct {
	SourceSymbolId uuid.UUID `json:"source_symbol_id"`
	Relation       string    `json:"relation"`
	DstSymbolId    uuid.UUID `json:"dst_symbol_id"`
	DetailsJson    string    `json:"details_json"`
}
