package main

import (
	"fmt"
	"log"
	tgbot "tgbot/audTextBot"

	"github.com/joho/godotenv"
)

// Init initializes the environment variables
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Sprint("Couldn't load local variables, err:", err))
	}
}

func main() {
	tgbot.StartBot()
}
