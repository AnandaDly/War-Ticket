package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"

	"github.com/joho/godotenv"

	"github.com/gofiber/fiber/v3"
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
	var stock int
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
		event = newEvent
		stock = newEvent.AvailableTickets
	} else {
		stock = event.AvailableTickets
	}

	err = rdb.Set(ctx, "ticket_stock", strconv.Itoa(stock), 0).Err()
	if err != nil {
		log.Fatal("Gagal set stok di Redis: ", err)
	} else {
		fmt.Println("Stok tiket berhasil disinkronisasi ke Redis!")
	}

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Ticket War API is running!")
	})

	app.Post("/buy", func(c fiber.Ctx) error {
		sisaTiket, err := rdb.Decr(ctx, "ticket_stock").Result()
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"message": "Stok tidak ditemukan"})
		}
		if sisaTiket < 0 {
			return c.Status(400).JSON(fiber.Map{
				"status":  "gagal",
				"message": "Maaf, tiket sudah habis!",
			})
		}
		return c.Status(200).JSON(fiber.Map{
			"status":     "sukses",
			"message":    "Berhasil mengamankan tiket!",
			"sisa_tiket": sisaTiket,
		})
	})

	log.Fatal(app.Listen(":8080"))
}
