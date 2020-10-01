package log

import (
	"fmt"
	"log"
	"strings"
)

func Infoln(v ...interface{}) {
	log.Output(2, fmt.Sprintf("\u001B[32m%s\u001B[0m", v...))
}

func Info(v ...interface{})  {
	log.Output(2, fmt.Sprint(v...))
}

func Infof(format string, v ...interface{}) {
	sprintf := fmt.Sprintf("\u001B[32m%s\u001B[0m", v...)
	strings.TrimSuffix(sprintf, "\n")
	log.Output(2, fmt.Sprintf(fmt.Sprintf("\u001B[32m%s\u001B[0m", format), v...))
}

func Errorln(v ...interface{})  {
	log.Output(2, fmt.Sprintf("\u001B[31m%s\u001B[0m", v...))
}

func Error(v ...interface{})  {
	log.Output(2, fmt.Sprint(fmt.Sprintf("\u001B[31m%s\u001B[0m", v...)))
}

func Errorf(format string, v ...interface{}) {
	log.Output(2, fmt.Sprintf(fmt.Sprintf("\u001B[31m%s\u001B[0m", format), v...))
}
