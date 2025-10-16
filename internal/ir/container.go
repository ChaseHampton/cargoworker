package ir

import "github.com/google/uuid"

type Container struct {
	Id         uuid.UUID `json:"id"`
	ProjectId  uuid.UUID `json:"project_id"`
	Language   string    `json:"language"`
	Name       string    `json:"name"`
	FullName   string    `json:"full_name"`
	Kind       string    `json:"kind"`
	VersionTag string    `json:"version_tag"`
	DocRaw     string    `json:"doc_raw"`
	DocFmt     string    `json:"doc_fmt"`
	ExtraJson  string    `json:"extra_json"`
}
