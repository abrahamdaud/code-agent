//go:build windows
package term

import (
	"unsafe"
	"syscall"
)

var (
	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode        = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode        = kernel32.NewProc("SetConsoleMode")
	procGetStdHandle          = kernel32.NewProc("GetStdHandle")
)

const (
	ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
	STD_OUTPUT_HANDLE                   = uint32(-11)
)

func Init() error {
	h, _, err := procGetStdHandle.Call(uintptr(STD_OUTPUT_HANDLE))
	if h == 0 {
		// nothing to do
		return nil
	}
	var mode uint32
	r, _, e := procGetConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode)))
	if r == 0 {
		_ = e
		return nil
	}
	mode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING
	procSetConsoleMode.Call(h, uintptr(mode))
	return nil
}
