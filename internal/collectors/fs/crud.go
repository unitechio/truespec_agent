package fs

import (
	"io"
	"os"
)

func Delete(path string) error {
	return os.Remove(path)
}

func Move(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	return copyAndDelete(src, dst)
}

func copyAndDelete(src, dst string) error {
	if err := copyFile(src, dst); err != nil {
		return err

	}
	return os.Remove(src)
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
}
