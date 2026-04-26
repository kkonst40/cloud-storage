package resource

import (
	"io"
)

type Response struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}

type FilePayload struct {
	Name    string
	Size    int64
	Content io.Reader
}
