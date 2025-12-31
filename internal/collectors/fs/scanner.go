package fs

import (
	"context"
	"io/fs"
	"path/filepath"
)

func ScanOS(ctx context.Context, out chan<- Node) error {
	defer close(out)

	for _, root := range getRoots() {
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // ignore permission
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			var size int64
			if !d.IsDir() {
				if info, err := d.Info(); err == nil {
					size = info.Size()
				}
			}

			out <- Node{
				Path:  path,
				Name:  d.Name(),
				IsDir: d.IsDir(),
				Size:  size,
				Type:  d.Type().String(),
			}
			return nil
		})
	}
	return nil
}
