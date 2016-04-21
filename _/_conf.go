package adb

import (
	"encoding/json"
	"syscall"
)

var (
	SYS_PAGE = syscall.Getpagesize()
	ENC      = json.Marshal
	DEC      = json.Unmarshal
)
