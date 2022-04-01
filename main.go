package main

import (
	"log"
	"os"
	paxfulMs "paxful/src"

	"github.com/joho/godotenv"
)

func main() {
	_ = os.Setenv("KE", "Africa/Nairobi")
	if err := godotenv.Load(); err != nil {
		log.Printf("unable to read dotenv %v", err)
	}
	paxfulMs.Run()
}
