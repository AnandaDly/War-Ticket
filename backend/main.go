package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `json:"name"`
	TicketTiers []TicketTier `gorm:"foreignKey:EventID" json:"ticket_tiers"`
}

type TicketTier struct {
	ID               uint   `json:"id"`
	EventID          uint   `json:"event_id"`
	Name             string `json:"name"`
	Price            int    `json:"price"`
	TotalTickets     int    `json:"total_tickets"`
	AvailableTickets int    `json:"available_tickets"`
}

type TicketOrder struct {
	ID           uint   `json:"id"`
	EventID      uint   `json:"event_id"`
	TicketTierID uint   `json:"ticket_tier_id"`
	UserID       string `json:"user_id"`
	Status       string `json:"status"`
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

type BuyRequest struct {
	EventID      uint `json:"event_id"`
	TicketTierID uint `json:"ticket_tier_id"`
}

type CreateEventRequest struct {
	Name        string       `json:"name"`
	TicketTiers []TicketTier `json:"ticket_tiers"`
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

func AuthRequired(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"message": "Akses ditolak, token tidak ada"})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(401).JSON(fiber.Map{"message": "Format token tidak valid"})
	}
	tokenString := parts[1]

	jwtSecret := os.Getenv("JWT_SECRET")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"message": "Token tidak valid atau kadaluarsa"})
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		c.Locals("user_id", claims["user_id"])
		c.Locals("role", claims["role"])
	}

	return c.Next()
}

func AdminOnly(c fiber.Ctx) error {
	roleVal := c.Locals("role")
	role, ok := roleVal.(string)
	if !ok || role != "admin" {
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak, hanya admin yang dapat mengakses"})
	}
	return c.Next()
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

	db.AutoMigrate(&Event{}, &TicketTier{}, &TicketOrder{}, &User{})

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
	err = db.Preload("TicketTiers").First(&event).Error
	if err != nil {
		newEvent := Event{
			Name: "Konser Coldplay",
			TicketTiers: []TicketTier{
				{Name: "VIP", Price: 5000000, TotalTickets: 50, AvailableTickets: 50},
				{Name: "Festival", Price: 2000000, TotalTickets: 150, AvailableTickets: 150},
			},
		}
		err = db.Create(&newEvent).Error
		if err != nil {
			log.Fatal("Gagal membuat event: ", err)
		} else {
			fmt.Println("Event berhasil dibuat: ", newEvent.Name)
		}
		event = newEvent
	}
	for _, tier := range event.TicketTiers {
		redisKey := fmt.Sprintf("event:%d:tier:%d:stock", event.ID, tier.ID)
		err = rdb.Set(ctx, redisKey, tier.AvailableTickets, 0).Err()
		if err != nil {
			log.Fatalf("Gagal set stok di Redis untuk tier %s: %v", tier.Name, err)
		} else {
			fmt.Printf("Stok tiket %s berhasil disinkronisasi ke Redis!\n", tier.Name)
		}
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

	app.Post("/buy", AuthRequired, func(c fiber.Ctx) error {
		var req BuyRequest
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"message": "Format data tidak valid"})
		}
		redisKey := fmt.Sprintf("event:%d:tier:%d:stock", req.EventID, req.TicketTierID)

		sisaTiket, err := rdb.Decr(ctx, redisKey).Result()
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"message": "Stok tidak ditemukan"})
		}
		if sisaTiket < 0 {
			return c.Status(400).JSON(fiber.Map{
				"status":  "gagal",
				"message": "Maaf, tiket sudah habis!",
			})
		}
		userVal := c.Locals("user_id")

		var UserIDStr string
		switch v := userVal.(type) {
		case float64:
			UserIDStr = fmt.Sprintf("%.0f", v)
		case int:
			UserIDStr = strconv.Itoa(v)
		case uint:
			UserIDStr = strconv.FormatUint(uint64(v), 10)
		case string:
			UserIDStr = v
		default:
			return c.Status(401).JSON(fiber.Map{"message": "ID User tidak valid atau tidak ditemukan"})
		}
		newOrder := TicketOrder{
			EventID:      req.EventID,
			TicketTierID: req.TicketTierID,
			UserID:       UserIDStr,
			Status:       "success",
		}
		orderChan <- newOrder
		return c.Status(200).JSON(fiber.Map{
			"status":     "sukses",
			"message":    "Berhasil mengamankan tiket!",
			"sisa_tiket": sisaTiket,
		})
	})

	// app.Get("/stock", func(c fiber.Ctx) error {
	// 	stok, err := rdb.Get(ctx, "ticket_stock").Result()
	// 	if err != nil {
	// 		if err == redis.Nil {
	// 			return c.Status(200).JSON(fiber.Map{"stock": 0})
	// 		}
	// 		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil data stok"})
	// 	}
	// 	return c.Status(200).JSON(fiber.Map{"stock": stok})
	// })

	app.Get("/events/:id", func(c fiber.Ctx) error {
		eventID := c.Params("id")
		var event Event

		if err := db.Preload("TicketTiers").First(&event, eventID).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"message": "Event tidak ditemukan"})
		}
		for i, tier := range event.TicketTiers {
			redisKey := fmt.Sprintf("event:%s:tier:%d:stock", eventID, tier.ID)

			stokStr, err := rdb.Get(ctx, redisKey).Result()
			if err == nil {
				stokInt, _ := strconv.Atoi(stokStr)
				event.TicketTiers[i].AvailableTickets = stokInt
			}
		}
		return c.Status(200).JSON(event)
	})

	app.Get("/events", func(c fiber.Ctx) error {
		var events []Event

		if err := db.Preload("TicketTiers").Find(&events).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil data event"})
		}

		return c.Status(200).JSON(events)
	})

	app.Post("/admin/events", AuthRequired, AdminOnly, func(c fiber.Ctx) error {
		var req CreateEventRequest
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"message": "Format data tidak valid"})
		}
		newEvent := Event{
			Name:        req.Name,
			TicketTiers: req.TicketTiers,
		}

		if err := db.Create(&newEvent).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal menyimpan event ke database"})
		}

		for _, tier := range newEvent.TicketTiers {
			redisKey := fmt.Sprintf("event:%d:tier:%d:stock", newEvent.ID, tier.ID)
			err := rdb.Set(ctx, redisKey, tier.TotalTickets, 0).Err()
			if err != nil {
				fmt.Println("❌ Gagal sinkronisasi Redis untuk tier:", tier.Name)
			}
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "Event berhasil dibuat dan disinkronisasi ke Redis!",
			"data":    newEvent,
		})
	})

	log.Fatal(app.Listen(":8080"))
}
