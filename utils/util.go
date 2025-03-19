package utils

import "runtime"

func IsWindowsOS() bool {
	os := runtime.GOOS
	if os == "windows" {
		return true
	}
	return false
}
