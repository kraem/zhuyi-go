package log

import (
	"log"
	"runtime"
)

func LogError(err error) (b bool) {
	if err != nil {
		pc, fn, line, _ := runtime.Caller(1)

		log.Printf("[error] [%s:%d] [%s] err: %v", fn, line, runtime.FuncForPC(pc).Name(), err)
		b = true
	}
	return
}
