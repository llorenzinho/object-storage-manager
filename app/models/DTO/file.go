package dto

type File struct {
	Name        string `json:"name"`
	Bucket      string `json:"bucket"`
	Verified    bool   `json:"verified"`
	ContentType string `json:"contentType,omitempty"`
	Content     []byte `json:"-"`
}

type FileList []File
