package types

import (
	"log"
)

type Loggers struct {
	ErrorLogger *log.Logger
	InfoLogger  *log.Logger
}
