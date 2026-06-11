package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/golang-jwt/jwt/v5"

	"github.com/redis/go-redis/v9"

	"github.com/joho/godotenv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
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

type User struct {
	ID       uint
	Name     string
	Email    string `gorm:"unique"`
	Password string
	Role     string
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func ticketWorker(orderChan chan TicketOrder, db *gorm.DB) {
	for order := range orderChan {
		err := db.Create(&order).Error
		if err != nil {
			fmt.Println("Gagal menyimpan tiket ke DB: ", err)
		} else {
			fmt.Println("Tiket tersimpan di DB untuk user: ", order.UserID)
		}
	}
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

	db.AutoMigrate(&Event{}, &TicketOrder{}, &User{})

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

	orderChan := make(chan TicketOrder, 100)
	go ticketWorker(orderChan, db)

	app := fiber.New()
	app.Use(cors.New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Ticket War API is running!")
	})

	app.Post("/register", func(c fiber.Ctx) error {
		var req RegisterRequest
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"message": "Data tidak valid"})
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal mengenkripsi password"})
		}
		newUser := User{
			Name:     req.Name,
			Email:    req.Email,
			Password: string(hashedPassword),
			Role:     "buyer",
		}
		var existingUser User
		err = db.Where("email = ?", req.Email).First(&existingUser).Error
		if err == nil {
			return c.Status(400).JSON(fiber.Map{"message": "Email sudah terdaftar"})
		}
		err = db.Create(&newUser).Error
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal membuat user"})
		}
		return c.Status(201).JSON(fiber.Map{"message": "Registrasi Berhasil"})
	})

	app.Post("/login", func(c fiber.Ctx) error {
		var req LoginRequest
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"message": "Data tidak valid"})
		}
		var user User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			return c.Status(401).JSON(fiber.Map{"message": "Email atau password salah"})
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			return c.Status(401).JSON(fiber.Map{"message": "Email atau password salah"})
		}

		claims := jwt.MapClaims{
			"user_id": user.ID,
			"email":   user.Email,
			"role":    user.Role,
			"exp":     time.Now().Add(time.Hour * 72).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		jwtSecret := os.Getenv("JWT_SECRET")
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal membuat token"})
		}
		return c.Status(200).JSON(fiber.Map{
			"message": "Login Berhasil",
			"token":   tokenString,
		})
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
		newOrder := TicketOrder{
			EventID: 1,
			UserID:  "user_pbl",
			Status:  "success",
		}
		orderChan <- newOrder
		return c.Status(200).JSON(fiber.Map{
			"status":     "sukses",
			"message":    "Berhasil mengamankan tiket!",
			"sisa_tiket": sisaTiket,
		})
	})

	app.Get("/stock", func(c fiber.Ctx) error {
		stok, err := rdb.Get(ctx, "ticket_stock").Result()
		if err != nil {
			if err == redis.Nil {
				return c.Status(200).JSON(fiber.Map{"stock": 0})
			}
			return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil data stok"})
		}
		return c.Status(200).JSON(fiber.Map{"stock": stok})
	})

	log.Fatal(app.Listen(":8080"))
}
