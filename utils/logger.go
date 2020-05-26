package utils

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	DEBUG = 36
	INFO  = 38
	WARN  = 34
	ERROR = 31
)

var level = map[int]string{
	DEBUG: "DBG",
	INFO:  "INF",
	WARN:  "WRN",
	ERROR: "ERR",
}

func ThrowCheck(err error) {
	if err != nil {
		Error(err)
		os.Exit(1)
	}
}

func Debug(vs ...interface{}) {
	print(DEBUG, vs)
}

func Info(vs ...interface{}) {
	print(INFO, vs)
}

func Warn(vs ...interface{}) {
	print(WARN, vs)
}

func Error(vs ...interface{}) {
	print(ERROR, vs)
}

func print(code int, vs interface{}) {
	fmt.Print(fmt.Sprintf(
		"%c[%d;%d;%dm[%s][%s] ==> %s%c[0m\n",
		0x1B, 0 /*字体*/, 0 /*背景*/, code /*前景*/, time.Now().Format("2006-01-02 15:04:05.000"), level[code],
		strings.Trim(fmt.Sprint(vs), "[]"),
		0x1B),
	)
}
