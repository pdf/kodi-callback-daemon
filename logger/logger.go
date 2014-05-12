package logger

import (
	`fmt`
	`log`
)

var DebugEnabled bool

func Debug(a ...interface{}) {
	if DebugEnabled == true {
		log.Println(`[DEBUG]`, fmt.Sprint(a...))
	}
}

func Info(a ...interface{}) {
	log.Println(`[INFO]`, fmt.Sprint(a...))
}

func Warn(a ...interface{}) {
	log.Println(`[WARN]`, fmt.Sprint(a...))
}

func Error(a ...interface{}) {
	log.Println(`[ERROR]`, fmt.Sprint(a...))
}

func Panic(a ...interface{}) {
	log.Panicln(`[PANIC]`, fmt.Sprint(a...))
}
