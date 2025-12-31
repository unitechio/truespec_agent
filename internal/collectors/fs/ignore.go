package fs

import "strings"

func shouldIgnore(path string) bool {
	p := strings.ToLower(path)

	ignore := []string{
		"/proc", "/sys", "/dev",
		"c:\\windows",
		"system volume information",
	}

	for _, ig := range ignore {
		if strings.Contains(p, ig) {
			return true
		}
	}
	return false
}
