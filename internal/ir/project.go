package ir

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	RootUri     string    `json:"root_uri"`
	ToolVersion string    `json:"tool_version"`
	IrSchema    string    `json:"ir_schema"`
	CreatedUtc  time.Time `json:"created_utc"`
}
