package adb

import (
	"os"
	"syscall"
	"unsafe"
)

type Data []byte

func Mmap(f *os.File, off, len int) Data {
	data, err := syscall.Mmap(int(f.Fd()), int64(off), len, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		panic(err)
	}
	return data
}

func (d Data) Mlock() {
	err := syscall.Mlock(d)
	if err != nil {
		panic(err)
	}
}

func (d Data) Munlock() {
	err := syscall.Munlock(d)
	if err != nil {
		panic(err)
	}
}

func (d Data) Munmap() {
	err := syscall.Munmap(d)
	d = nil
	if err != nil {
		panic(err)
	}
}

func (d Data) Sync() {
	_, _, err := syscall.Syscall(syscall.SYS_MSYNC,
		uintptr(unsafe.Pointer(&d[0])), uintptr(len(d)),
		uintptr(syscall.MS_ASYNC))
	if err != 0 {
		panic(err)
	}
}
