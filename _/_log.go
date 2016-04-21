package adb

import "log"

func Logger(str string, args ...interface{}) {
	log.Printf(str, args...)
}
