package audTextBot

import (
	"log"
	"os"
	"runtime"

	service "tgbot/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	redisService        = service.NewRedisService()
	UserCallBackCh      = make(chan string)
	allowedAudioFormats = []string{".3ga", ".8svx", ".aac", ".ac3", ".aif", ".aiff", ".alac", ".amr", ".ape", ".au", ".dss", ".flac", ".flv", ".m4a", ".m4b", ".m4p", ".m4r", ".mp3", ".mpga", ".ogg", ".oga", ".mogg", ".opus", ".qcp", ".tta", ".vocn", ".wavn", ".wma", ".wv"}
	commandsMap         = map[string]func(*tgbotapi.BotAPI, tgbotapi.Update){
		"/start":      handleStart,
		"/premium":    handlePremium,
		"/list":       handleList,
		"/getPremium": getPremium,
		"/status":     checkStatus,
	}
)

func StartBot() {
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
				go handleUpdate(bot, update)
			}
		}()
	}
	for {
	}
}

func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
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
		handleAudio(bot, update)
	case update.Message.Voice != nil:
		handleVoice(bot, update)
	default:
		if command, ok := commandsMap[update.Message.Text]; ok {
			command(bot, update)
		} else {
			sendMessage(bot, update.Message.Chat.ID, "Invalid message")
		}
	}
}
