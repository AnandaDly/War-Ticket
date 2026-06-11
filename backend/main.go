package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"

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

	REDIS_URL := os.Getenv("REDIS_URL")
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     REDIS_URL,
		Password: "",
		DB:       0,
	})

	err = rdb.Ping(ctx).Err()
	if err != nil {
		panic(err)
	} else {
		fmt.Println("berhasil connect ke redis")
	}

	var event Event
	err = db.First(&event).Error
	if err != nil {
		newEvent := Event{
			Name:             "Konser Coldplay",
			TotalTickets:     100,
			AvailableTickets: 100,
		}
		err = db.Create(&newEvent).Error
		if err != nil {
			log.Fatal("Gagal membuat event: ", err)
		} else {
			fmt.Println("Event berhasil dibuat: ", newEvent.Name)
		}
	}
}
