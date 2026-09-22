package utl

import (
	"fmt"
	"os"
	"time"

	//    "io"
	//    "io/ioutil"
	"bufio"
)

/*
LogLevel
- DEBUG     10
- INFO      20
- WARN      30
- ERROR     40
*/
var GLOGLEVEL string = "DEBUG"
var logFileName = "log.txt"

func Log(logLevel string, msg string) {

	dt := string(time.Now().Format("2006/01/02 15:04:05"))

	// GLOGLEVEL 为 DEBUG 时，所有日志额外写入文件
	if GLOGLEVEL == "DEBUG" {
		logMess := fmt.Sprintf("[%s %s] %s\n", dt, logLevel, msg)
		file, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			// 不能在 Log 内部再调用 Log（会递归），错误直接输出到 stderr
			fmt.Fprintf(os.Stderr, "Error in open log file [%s], error = %v\n", logFileName, err)
		} else {
			defer file.Close()
			write := bufio.NewWriter(file)
			write.WriteString(logMess)
			write.Flush()
		}
	}

	// logLevel: DEBUG INFO WARN ERROR
	// 所有级别均输出到终端
	fmt.Printf("[\x1b[47;30m%s\x1b[0m\x1b[43;30m%8s\x1b[0m] %s\n", dt, logLevel, msg)
}
