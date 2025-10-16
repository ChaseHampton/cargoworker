package ir

import "github.com/google/uuid"

type Member struct {
	Id            uuid.UUID `json:"id"`
	OwnerSymbolId uuid.UUID `json:"owner_symbol_id"`
	ChildSymbolId uuid.UUID `json:"child_symbol_id"`
	Order         int       `json:"order"`
}
