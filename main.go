package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Hey go!")

	err := godotenv.Load()

	if err != nil {
		log.Fatal("error loading env file")
	}

	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello Fibre")
	})

	log.Fatal(app.Listen(":8080"))

}
