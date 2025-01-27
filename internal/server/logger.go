package server

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var LogFileEnabled bool

func InitLog(opts *Options) {
	if opts.Log == "" {
		return
	}
	LogFileEnabled = true

	handle, err := os.OpenFile(opts.Log, os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	logger := &logWriter{
		LogHandle: handle,
	}
	log.SetOutput(logger)
	log.Println("< < logging start > >")
}

type logWriter struct {
	LogHandle *os.File
}

func (l *logWriter) Write(p []byte) (int, error) {
	if !strings.Contains(string(p), "< < logging start > >") {
		fmt.Print(string(p))
	}
	return l.LogHandle.Write(p)
}
