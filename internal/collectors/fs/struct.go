package fs

type Node struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	Type    string
	Mode    uint32
	ModTime int64

	Children []*Node `json:",omitempty"`
}

type FileIndex struct {
	Path    string
	Parent  string
	Name    string
	IsDir   bool
	Size    int64
	ModTime int64
}
