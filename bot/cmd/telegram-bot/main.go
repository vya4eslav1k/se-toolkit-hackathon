package main

import (
	"fmt"
	"log"
	"os"

	"github.com/maksimslavik/inno-se-toolkit-pet/bot/internal"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	apiURL := os.Getenv("BACKEND_API_URL")
	if apiURL == "" {
		apiURL = "http://backend:8080/api"
	}

	b, err := bot.New(token, apiURL)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	fmt.Println("Starting Telegram Bot...")
	b.Start()
}
