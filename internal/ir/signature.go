package ir

import "github.com/google/uuid"

type Signature struct {
	SymbolId uuid.UUID `json:"symbol_id"`
	Text     string    `json:"text"`
	Json     string    `json:"json"`
}
