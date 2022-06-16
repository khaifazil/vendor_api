package main

import (
	"io"
	"log"
	"os"
)

var errLog *os.File
var ErrorLogger *log.Logger

func init() {
	var err error

	errLog, err = os.OpenFile("logs/errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	flags := log.LstdFlags | log.Lshortfile
	ErrorLogger = log.New(io.MultiWriter(errLog, os.Stderr), "ERROR: ", flags)
}
