package fs

import (
	"io/fs"
	"path/filepath"
)

func BuildFSTree(root string) (*Node, error) {
	rootNode := &Node{
		Name:  filepath.Base(root),
		Path:  root,
		IsDir: true,
	}

	nodeMap := map[string]*Node{
		root: rootNode,
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip permission denied
		}

		if path == root {
			return nil
		}

		parent := filepath.Dir(path)
		parentNode, ok := nodeMap[parent]
		if !ok {
			return nil
		}

		info, _ := d.Info()

		node := &Node{
			Name:    d.Name(),
			Path:    path,
			IsDir:   d.IsDir(),
			Size:    info.Size(),
			Mode:    uint32(info.Mode()),
			ModTime: info.ModTime().Unix(),
		}

		parentNode.Children = append(parentNode.Children, node)

		if d.IsDir() {
			nodeMap[path] = node
		}

		return nil
	})

	return rootNode, err
}
