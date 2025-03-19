package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	fmt.Println("Hey go!")
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello Fibre")
	})

	log.Fatal(app.Listen(":8080"))

}
