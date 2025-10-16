package ir

import "github.com/google/uuid"

type Typeref struct {
	Id            uuid.UUID `json:"id"`
	OwnerSymbolId uuid.UUID `json:"owner_symbol_id"`
	Slot          string    `json:"slot"`
	Json          string    `json:"json"`
	Order         int       `json:"order"`
}
