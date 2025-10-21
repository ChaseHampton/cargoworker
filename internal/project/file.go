package project

type FileMeta struct {
	Path  string `json:"path"`
	Depth int    `json:"depth"`
	IsDir bool   `json:"is_dir"`
}
