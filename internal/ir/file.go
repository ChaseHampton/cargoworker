package ir

import "github.com/google/uuid"

type File struct {
	Id        uuid.UUID `json:"id"`
	ProjectId uuid.UUID `json:"project_id"`
	Path      string    `json:"path"`
	Checksum  string    `json:"checksum"`
	Language  string    `json:"language"`
}
