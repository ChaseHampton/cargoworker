package ir

import "github.com/google/uuid"

type Diagnostic struct {
	Id       uuid.UUID `json:"id"`
	Scope    string    `json:"scope"`
	Severity string    `json:"severity"`
	Code     string    `json:"code"`
	Message  string    `json:"message"`
	FileId   uuid.UUID `json:"file_id"`
	Line     int       `json:"line"`
	Column   int       `json:"column"`
}
