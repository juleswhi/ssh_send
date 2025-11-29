package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/juleswhi/yoshimi/pkg/ssh"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Print(err)
	}
    ssh.Init()
}
