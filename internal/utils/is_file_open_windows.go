//go:build windows

package utils

import (
	"syscall"
)

// IsFileOpen checks if a file is currently open by any process
func IsFileOpen(path string) bool {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false
	}

	// On Windows, try to open with exclusive access using syscall
	handle, err := syscall.CreateFile(
		ptr,
		syscall.GENERIC_READ,
		0, // No sharing - will fail if file is open
		nil,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return true // File is open or otherwise inaccessible
	}
	syscall.CloseHandle(handle)
	return false
}
