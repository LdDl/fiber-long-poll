package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	lp "github.com/LdDl/fiber-long-poll/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/valyala/fasthttp"
)

var (
	userID = "my_pretty_uuid"
)

func main() {

	config := fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			log.Println(err)
			return ctx.Status(fasthttp.StatusInternalServerError).JSON(errors.New("panic error"))
		},
		IdleTimeout: 10 * time.Second,
	}
	allCors := cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Authorization, Content-Type, Content-Length, Accept, Accept-Encoding, X-HttpRequest",
		AllowMethods:     "GET, POST, PUT, DELETE",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
		MaxAge:           5600,
	})

	server := fiber.New(config)
	server.Use(allCors)

	manager, err := lp.StartLongpoll(lp.Options{
		LoggingEnabled:                 false,
		MaxLongpollTimeoutSeconds:      120,
		MaxEventBufferSize:             100,
		EventTimeToLiveSeconds:         60 * 2,
		DeleteEventAfterFirstRetrieval: false,
	})
	if err != nil {
		log.Printf("Failed to create manager: %q", err)
		return
	}
	defer manager.Shutdown()

	go generatingMessages(manager)
	server.Get("/unread_messages", GetMessages(manager))

	err = server.Listen(":8080")
	if err != nil {
		fmt.Printf("Can't start server due the error: %s\n", err.Error())
	}
}

// GetMessages Long polling request
func GetMessages(manager *lp.LongpollManager) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		ctx.Context().PostArgs().Set("timeout", "10")
		ctx.Context().PostArgs().Set("category", fmt.Sprintf("unread_messages_for_%s", userID))
		return manager.SubscriptionHandler(ctx)
	}
}

// generatingMessages Generate some messages
func generatingMessages(manager *lp.LongpollManager) {
	i := 0
	for {
		manager.Publish(fmt.Sprintf("unread_messages_for_%s", userID), fmt.Sprintf("Number: %d", i))
		i++
		time.Sleep(3 * time.Second)
	}
}
