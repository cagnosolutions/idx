package idx

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

func (d Data) Mremap(size int) Data {
	fd := uintptr(unsafe.Pointer(&d[0]))
	err := syscall.Munmap(d)
	d = nil
	if err != nil {
		panic(err)
	}
	err = syscall.Ftruncate(int(fd), int64(align(size)))
	if err != nil {
		panic(err)
	}
	d, err = syscall.Mmap(int(fd), int64(0), size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		panic(err)
	}
	return d
}

// open file helper
func OpenFile(path string) (*os.File, string, int) {
	fd, err := os.OpenFile(path, syscall.O_RDWR|syscall.O_CREAT|syscall.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	fi, err := fd.Stat()
	if err != nil {
		panic(err)
	}
	return fd, sanitize(fi.Name()), int(fi.Size())
}

// round up to nearest pagesize -- helper
func align(size int) int {
	if size > 0 {
		return (size + SYS_PAGE - 1) &^ (SYS_PAGE - 1)
	}
	return SYS_PAGE
}

// resize underlying file -- helper
func resize(fd uintptr, size int) int {
	err := syscall.Ftruncate(int(fd), int64(align(size)))
	if err != nil {
		panic(err)
	}
	return size
}
