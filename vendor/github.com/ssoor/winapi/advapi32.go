// Copyright 2010 The win Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"
)

const KEY_READ REGSAM = 0x20019
const KEY_WRITE REGSAM = 0x20006

const (
	HKEY_CLASSES_ROOT     HKEY = 0x80000000
	HKEY_CURRENT_USER     HKEY = 0x80000001
	HKEY_LOCAL_MACHINE    HKEY = 0x80000002
	HKEY_USERS            HKEY = 0x80000003
	HKEY_PERFORMANCE_DATA HKEY = 0x80000004
	HKEY_CURRENT_CONFIG   HKEY = 0x80000005
	HKEY_DYN_DATA         HKEY = 0x80000006
)

const (
	ERROR_NO_MORE_ITEMS = 259
)

type (
	ACCESS_MASK uint32
	HKEY        HANDLE
	REGSAM      ACCESS_MASK
)

const (
	REG_NONE      uint64 = 0 // No value type
	REG_SZ               = 1 // Unicode nul terminated string
	REG_EXPAND_SZ        = 2 // Unicode nul terminated string
	// (with environment variable references)
	REG_BINARY                     = 3 // Free form binary
	REG_DWORD                      = 4 // 32-bit number
	REG_DWORD_LITTLE_ENDIAN        = 4 // 32-bit number (same as REG_DWORD)
	REG_DWORD_BIG_ENDIAN           = 5 // 32-bit number
	REG_LINK                       = 6 // Symbolic Link (unicode)
	REG_MULTI_SZ                   = 7 // Multiple Unicode strings
	REG_RESOURCE_LIST              = 8 // Resource list in the resource map
	REG_FULL_RESOURCE_DESCRIPTOR   = 9 // Resource list in the hardware description
	REG_RESOURCE_REQUIREMENTS_LIST = 10
	REG_QWORD                      = 11 // 64-bit number
	REG_QWORD_LITTLE_ENDIAN        = 11 // 64-bit number (same as REG_QWORD)

)

const (
	REG_OPTION_NON_VOLATILE = 0

	REG_CREATED_NEW_KEY     = 1
	REG_OPENED_EXISTING_KEY = 2
)

var (
	// Library
	libadvapi32 uintptr

	// Functions
	regCloseKey     uintptr
	regOpenKeyEx    uintptr
	regDeleteKey    uintptr
	regCreateKeyEx  uintptr
	regQueryValueEx uintptr
	regEnumValue    uintptr
	regSetValueEx   uintptr
)

func init() {
	// Library
	libadvapi32 = MustLoadLibrary("advapi32.dll")

	// Functions
	regCloseKey = MustGetProcAddress(libadvapi32, "RegCloseKey")
	regOpenKeyEx = MustGetProcAddress(libadvapi32, "RegOpenKeyExW")
	regDeleteKey = MustGetProcAddress(libadvapi32, "RegDeleteKeyW")
	regCreateKeyEx = MustGetProcAddress(libadvapi32, "RegCreateKeyExW")
	regQueryValueEx = MustGetProcAddress(libadvapi32, "RegQueryValueExW")
	regEnumValue = MustGetProcAddress(libadvapi32, "RegEnumValueW")
	regSetValueEx = MustGetProcAddress(libadvapi32, "RegSetValueExW")
}

func RegDeleteKey(hKey HKEY, subkey string) (regerrno error) {
	ret, _, _ := syscall.Syscall(regDeleteKey, 2,
		uintptr(hKey),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(subkey))), 
		0)
		
	if ret != 0 {
		regerrno = syscall.Errno(ret)
	}
	return
}

func RegCreateKeyEx(hKey HKEY, subkey string, reserved uint32, class *uint16, options uint32, desired uint32, sa *syscall.SecurityAttributes, result *HKEY, disposition *uint32) (regerrno error) {
	ret, _, _ := syscall.Syscall9(regCreateKeyEx, 9,
		uintptr(hKey),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(subkey))),
		uintptr(reserved),
		uintptr(unsafe.Pointer(class)),
		uintptr(options),
		uintptr(desired),
		uintptr(unsafe.Pointer(sa)),
		uintptr(unsafe.Pointer(result)),
		uintptr(unsafe.Pointer(disposition)))

	if ret != 0 {
		regerrno = syscall.Errno(ret)
	}
	return
}

func RegCloseKey(hKey HKEY) int32 {
	ret, _, _ := syscall.Syscall(regCloseKey, 1,
		uintptr(hKey),
		0,
		0)

	return int32(ret)
}

func RegOpenKeyEx(hKey HKEY, lpSubKey string, ulOptions uint32, samDesired REGSAM, phkResult *HKEY) int32 {
	ret, _, _ := syscall.Syscall6(regOpenKeyEx, 5,
		uintptr(hKey),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpSubKey))),
		uintptr(ulOptions),
		uintptr(samDesired),
		uintptr(unsafe.Pointer(phkResult)),
		0)

	return int32(ret)
}

func RegQueryValueEx(hKey HKEY, lpValueName string, lpReserved uint32, lpType *uint32, lpData *[]uint16, lpcbData *uint32) int32 {
	ret, _, _ := syscall.Syscall6(regQueryValueEx, 6,
		uintptr(hKey),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpValueName))),
		uintptr(lpReserved),
		uintptr(unsafe.Pointer(lpType)),
		uintptr(unsafe.Pointer(&(*lpData)[0])),
		uintptr(unsafe.Pointer(lpcbData)))

	return int32(ret)
}

func RegEnumValue(hKey HKEY, index uint32, lpValueName string, lpcchValueName *uint32, lpReserved uint32, lpType *uint32, lpData *byte, lpcbData *uint32) int32 {
	ret, _, _ := syscall.Syscall9(regEnumValue, 8,
		uintptr(hKey),
		uintptr(index),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpValueName))),
		uintptr(unsafe.Pointer(lpcchValueName)),
		uintptr(lpReserved),
		uintptr(unsafe.Pointer(lpType)),
		uintptr(unsafe.Pointer(lpData)),
		uintptr(unsafe.Pointer(lpcbData)),
		0)
	return int32(ret)
}

func RegSetValueEx(hKey HKEY, lpValueName string, lpReserved uint32, lpDataType uint32, lpData *byte, cbData uint32) int32 {
	ret, _, _ := syscall.Syscall6(regSetValueEx, 6,
		uintptr(hKey),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpValueName))),
		uintptr(lpReserved),
		uintptr(lpDataType),
		uintptr(unsafe.Pointer(lpData)),
		uintptr(cbData))
	return int32(ret)
}
