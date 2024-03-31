package log

import (
	"io"
	"log"
	"os"
)

var (

	// Trace Just about anything

	Trace *log.Logger

	// Info Important information

	Info *log.Logger

	// Warning Be concerned

	Warning *log.Logger

	// Error Critical problem

	Error *log.Logger
)

// Initialize will initialize logger

func Initialize() {
	Trace = log.New(os.Stdout,

		"TRACE: ",

		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(os.Stdout,

		"INFO: ",

		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(os.Stdout,

		"WARNING: ",

		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(io.MultiWriter(os.Stderr),

		"ERROR: ",

		log.Ldate|log.Ltime|log.Lshortfile)
}
