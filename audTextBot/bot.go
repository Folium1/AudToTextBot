package audTextBot

import (
	"log"
	"os"
	"runtime"

	service "tgbot/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

var (
	UserCallBackCh      = make(chan string)
	allowedAudioFormats = []string{".3ga", ".8svx", ".aac", ".ac3", ".aif", ".aiff", ".alac", ".amr", ".ape", ".au", ".dss", ".flac", ".flv", ".m4a", ".m4b", ".m4p", ".m4r", ".mp3", ".mpga", ".ogg", ".oga", ".mogg", ".opus", ".qcp", ".tta", ".vocn", ".wavn", ".wma", ".wv"}
	commandsMap         = map[string]func(*tgbotapi.BotAPI, tgbotapi.Update, *service.RedisService){
		"/start":      handleStart,
		"/premium":    handlePremium,
		"/list":       handleList,
		"/getPremium": getPremium,
	}
)

func StartBot() {
	redisService, err := service.NewRedisService()
	if err != nil {
		log.Println(err)
	}
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	botToken := os.Getenv("BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	// Create a channel to receive updates from the updates channel
	updateChannel := make(chan tgbotapi.Update, runtime.NumCPU())

	// Launch a goroutine to listen for updates and send them to the update channel
	go func() {
		for update := range updates {
			updateChannel <- update
		}
	}()

	// Start multiple goroutines to handle updates concurrently
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for update := range updateChannel {
				go handleUpdate(bot, update, redisService)
			}
		}()
	}
	for {
	}
}

func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update, redisService *service.RedisService) {
	msg := tgbotapi.MessageConfig{}
	if update.Message != nil {
		msg.ChatID = update.Message.Chat.ID
	}
	switch {
	case update.CallbackQuery != nil:
		UserCallBackCh <- update.CallbackQuery.Data
	case update.Message == nil:
		return
	case update.Message.Audio != nil:
		handleAudio(bot, update, update.Message.Audio, redisService)
	case update.Message.Voice != nil:
		handleVoice(bot, update, update.Message.Voice, redisService)
	default:
		if command, ok := commandsMap[update.Message.Text]; ok {
			command(bot, update, redisService)
		} else {
			sendMessage(bot, update.Message.Chat.ID, "Invalid message")
		}
	}
}
