package fs

import (
	"fmt"
	"os"
	"runtime"
)

func getRoots() []string {
	switch runtime.GOOS {
	case "windowns":
		return listWindownsDrivers()
	case "linux":
		return []string{"/"}
	case "darwin":
		return []string{"/"}
	default:
		return []string{"/"}
	}
}

func listWindownsDrivers() []string {
	drivers := []string{}
	for i := 'A'; i <= 'Z'; i++ {
		driver := fmt.Sprintf("%c:\\", i)
		if _, err := os.Stat(driver); err == nil {
			drivers = append(drivers, driver)
		}
	}
	return drivers
}
