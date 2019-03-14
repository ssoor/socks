package common

import (
	"fmt"
	"runtime"
	"syscall"

	"github.com/ssoor/winapi"
)

func GetCPUString() (string, error) {
	var regHKey winapi.HKEY
	if errorCode := winapi.RegOpenKeyEx(winapi.HKEY_LOCAL_MACHINE, "HARDWARE\\DESCRIPTION\\System\\CentralProcessor\\0", 0, winapi.KEY_READ, &regHKey); winapi.ERROR_SUCCESS != errorCode {
		return "", nil
	}

	var bufSize uint32 = 256
	bufCPUName := make([]uint16, bufSize)

	if errorCode := winapi.RegQueryValueEx(regHKey, "ProcessorNameString", 0, nil, &bufCPUName, &bufSize); winapi.ERROR_SUCCESS != errorCode {
		return "", nil
	}

	winapi.RegCloseKey(regHKey)

	return fmt.Sprintf("%s (%d)", syscall.UTF16ToString(bufCPUName), runtime.NumCPU()), nil
}
