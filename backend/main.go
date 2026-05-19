package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/joho/godotenv"
)

type Event struct {
	ID               uint
	Name             string
	TotalTickets     int
	AvailableTickets int
}

type TicketOrder struct {
	ID      uint
	EventID uint
	UserID  string
	Status  string
}

func main() {
	fmt.Println("Memulai Ticket War Engine...")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	DB_URL := os.Getenv("DB_URL")
	fmt.Println(DB_URL)

	db, err := gorm.Open(postgres.Open(DB_URL), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	} else {
		fmt.Println("Berhasil Connect ke database")
	}

	db.AutoMigrate(&Event{}, &TicketOrder{})
	// REDIS_URL := os.Getenv("REDIS_URL")

}
