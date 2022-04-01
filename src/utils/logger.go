package utils

import (
	"log"
	"os"

)

func InitLogger() {
	log.SetOutput(os.Stderr)
	return
}
