package main

import (
	"log"
	"news/internal/services"
	"news/internal/transport"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	//
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {

	file, err := os.OpenFile("news.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer file.Close()
	log.SetOutput(file)

	go services.Getrss()
	transport.Server()

}
