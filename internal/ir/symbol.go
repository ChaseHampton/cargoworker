package ir

import "github.com/google/uuid"

type Symbol struct {
	Id           uuid.UUID `json:"id"`
	ContainerId  uuid.UUID `json:"container_id"`
	Name         string    `json:"name"`
	FullName     string    `json:"full_name"`
	Kind         string    `json:"kind"`
	Visibility   string    `json:"visibility"`
	Flags        int       `json:"flags"`
	OriginFileId uuid.UUID `json:"origin_file_id"`
	StartLine    int       `json:"start_line"`
	StartCol     int       `json:"start_col"`
	EndLine      int       `json:"end_line"`
	EndCol       int       `json:"end_col"`
	DocRaw       string    `json:"doc_raw"`
	DocFmt       string    `json:"doc_fmt"`
	ExtraJson    string    `json:"extra_json"`
}
