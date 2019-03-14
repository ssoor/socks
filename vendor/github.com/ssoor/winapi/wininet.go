// Copyright 2011 The win Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"
)

//
// access types for InternetOpen()
//
const (
	INTERNET_OPEN_TYPE_PRECONFIG                   uint32 = 0 // use registry configuration
	INTERNET_OPEN_TYPE_DIRECT                             = 1 // direct to net
	INTERNET_OPEN_TYPE_PROXY                              = 3 // via named proxy
	INTERNET_OPEN_TYPE_PRECONFIG_WITH_NO_AUTOPROXY        = 4 // prevent using java/script/INS
)

//
// Options used in INTERNET_PER_CONN_OPTON struct
//
const (
	INTERNET_PER_CONN_FLAGS                        uint32 = 1
	INTERNET_PER_CONN_PROXY_SERVER                        = 2
	INTERNET_PER_CONN_PROXY_BYPASS                        = 3
	INTERNET_PER_CONN_AUTOCONFIG_URL                      = 4
	INTERNET_PER_CONN_AUTODISCOVERY_FLAGS                 = 5
	INTERNET_PER_CONN_AUTOCONFIG_SECONDARY_URL            = 6
	INTERNET_PER_CONN_AUTOCONFIG_RELOAD_DELAY_MINS        = 7
	INTERNET_PER_CONN_AUTOCONFIG_LAST_DETECT_TIME         = 8
	INTERNET_PER_CONN_AUTOCONFIG_LAST_DETECT_URL          = 9
	INTERNET_PER_CONN_FLAGS_UI                            = 10
)

//
// PER_CONN_FLAGS
//
const (
	PROXY_TYPE_DIRECT         uint64 = 0x00000001 // direct to net
	PROXY_TYPE_PROXY                 = 0x00000002 // via named proxy
	PROXY_TYPE_AUTO_PROXY_URL        = 0x00000004 // autoproxy URL
	PROXY_TYPE_AUTO_DETECT           = 0x00000008 // use autoproxy detection
)

const (
	INTERNET_OPTION_REFRESH uint32 = 37
	INTERNET_OPTION_PER_CONNECTION_OPTION uint32 = 75
)

type (
	HINTERNET *byte
)

//
// INTERNET_PROXY_INFO - structure supplied with INTERNET_OPTION_PROXY to get/
// set proxy information on a InternetOpen() handle
//
type INTERNET_PROXY_INFO struct {
	dwAccessType    uint32
	lpszProxy       *uint16
	lpszProxyBypass *uint16
}

/*
type FILETIME struct {
	dwLowDateTime  uint32
	dwHighDateTime uint32
}
*/
type INTERNET_PER_CONN_OPTION struct {
	Option uint32 // option to be queried or set // union
	Value  uint64 // union
}

type INTERNET_PER_CONN_OPTION_LIST struct {
	Size        uint32  // size of the INTERNET_PER_CONN_OPTION_LIST struct
	Connection  *uint16 // connection name to set/query options
	OptionCount uint32  // number of options to set/query
	OptionError uint32  // on error, which option failed
	Options     *INTERNET_PER_CONN_OPTION
	// array of options to set/query
}

var (
	// Library
	libwininet uintptr

	// Functions
	internetSetOption uintptr
)

func init() {
	// Library
	libwininet = MustLoadLibrary("wininet.dll")

	// Functions
	internetSetOption = MustGetProcAddress(libwininet, "InternetSetOptionW")
}

type InternetPtr interface{}

func InternetSetOption(hInternet HINTERNET, dwOption uint32, interfaceBuffer InternetPtr, dwBufferLength uint32) bool {
	var lpBuffer unsafe.Pointer

	switch dwOption {
	case INTERNET_OPTION_PER_CONNECTION_OPTION:
		lpBuffer = unsafe.Pointer(interfaceBuffer.(*INTERNET_PER_CONN_OPTION_LIST))
	}

	ret, _, _ := syscall.Syscall6(internetSetOption, 4,
		uintptr(unsafe.Pointer(hInternet)),
		uintptr(dwOption),
		uintptr(lpBuffer),
		uintptr(dwBufferLength),
		0,
		0)

	return (0 != ret)
}
