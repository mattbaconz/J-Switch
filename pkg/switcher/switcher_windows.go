//go:build windows

package switcher

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const (
	// HWND_BROADCAST is the handle for broadcasting to all top-level windows.
	HWND_BROADCAST = 0xffff
	// WM_SETTINGCHANGE is the message to indicate a system setting change.
	WM_SETTINGCHANGE = 0x001A
	// SMTO_ABORTIFHUNG returns without waiting for the time-out period to elapse if the receiving thread appears to not be in a state to receive the message.
	SMTO_ABORTIFHUNG = 0x0002
)

func switchJava(javaPath string) error {
	// 1. Open Registry Key HKCU\Environment
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()

	// 2. Set JAVA_HOME
	if err := k.SetStringValue("JAVA_HOME", javaPath); err != nil {
		return fmt.Errorf("failed to set JAVA_HOME: %w", err)
	}

	// 3. Update Path
	pathValue, _, err := k.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to read Path: %w", err)
	}

	javaHomeBin := `%JAVA_HOME%\bin`

	// Split path (usually simply with ;)
	// Note: Windows environment path separator is ;
	parts := strings.Split(pathValue, ";")

	// Clean up existing %JAVA_HOME%\bin if present to move/ensure it is at front?
	// Or just check if it is already at the front.
	// For robustness, let's filter it out and prepend it.

	var newParts []string
	newParts = append(newParts, javaHomeBin)

	for _, p := range parts {
		cleanP := strings.TrimSpace(p)
		if cleanP == "" {
			continue
		}
		// Check for %JAVA_HOME%\bin case-insensitive or resolved
		if strings.EqualFold(cleanP, javaHomeBin) {
			continue
		}
		newParts = append(newParts, cleanP)
	}

	newPathValue := strings.Join(newParts, ";")

	if err := k.SetStringValue("Path", newPathValue); err != nil {
		return fmt.Errorf("failed to set Path: %w", err)
	}

	// 4. Broadcast Change
	if err := broadcastEnvironmentChange(); err != nil {
		fmt.Printf("Warning: Failed to broadcast environment change: %v\n", err)
		fmt.Println("You may need to restart your terminal or log out/in.")
	} else {
		fmt.Println("Environment variables updated and broadcasted.")
	}

	return nil
}

func broadcastEnvironmentChange() error {
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessageTimeout := user32.NewProc("SendMessageTimeoutW")

	environmentPtr, _ := syscall.UTF16PtrFromString("Environment")

	// LRESULT SendMessageTimeoutW(
	//   HWND       hWnd,
	//   UINT       Msg,
	//   WPARAM     wParam,
	//   LPARAM     lParam,
	//   UINT       fuFlags,
	//   UINT       uTimeout,
	//   PDWORD_PTR lpdwResult
	// );

	ret, _, err := sendMessageTimeout.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(environmentPtr)),
		uintptr(SMTO_ABORTIFHUNG),
		uintptr(5000), // 5 seconds timeout
		0,
	)

	if ret == 0 {
		return fmt.Errorf("SendMessageTimeout failed: %v", err)
	}

	return nil
}
