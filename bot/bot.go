package bot

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"

	service "tgbot/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

const (
	unpremiumMaxAudioDuration = 480 // maximum allowed audio duration for non-premium users (in seconds)
)

var (
	allowedAudioFormats = []string{".3ga", ".8svx", ".aac", ".ac3", ".aif", ".aiff", ".alac", ".amr", ".ape", ".au", ".dss", ".flac", ".flv", ".m4a", ".m4b", ".m4p", ".m4r", ".mp3", ".mpga", ".ogg", ".oga", ".mogg", ".opus", ".qcp", ".tta", ".vocn", ".wavn", ".wma", ".wv"}
	commandsMap         = map[string]func(*tgbotapi.BotAPI, tgbotapi.Update, *service.RedisService){
		"/start":   handleStart,
		"/premium": handlePremium,
		"/list":    handleList,
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
	msg.ChatID = update.Message.Chat.ID
	switch {
	case update.Message == nil:
		return
	case update.Message.Audio != nil:
		handleAudio(bot, update, update.Message.Audio, redisService)
	case update.Message.Voice != nil:
		handleVoice(bot, update, update.Message.Voice, redisService)
	default:
		if command, ok := commandsMap[update.Message.Text]; ok {
			command(bot, update, redisService)
		}
	}

}

// sendMessage sends a message to a chat
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func handleAudio(bot *tgbotapi.BotAPI, update tgbotapi.Update, audio *tgbotapi.Audio, redisService *service.RedisService) {
	msg := tgbotapi.MessageConfig{}
	msg.ChatID = update.Message.Chat.ID

	// Check if user is premium
	isPremium := isPremium(bot, update.Message.From.ID, update.Message.Chat.ID, redisService)

	// Check if audio duration is allowed
	err := isAudioDurationAllowed(update.Message.Audio.Duration, update.Message.From.ID, update.Message.From.FirstName, isPremium, redisService)
	if err != nil {
		sendMessage(bot, update.Message.Chat.ID, err.Error())
		return
	}

	// Send initial message
	sendMessage(bot, update.Message.Chat.ID, fmt.Sprintf("Decoding will take from 15%% to 30%% of file duration if it is not too short"))

	// Decode audio file
	text, err := decodeAudioFile(bot, update.Message.Audio.FileID, redisService)
	if err != nil {
		log.Println(err)
		sendMessage(bot, update.Message.Chat.ID, "There was an error decoding the file")
		return
	}
	sendMessage(bot, update.Message.Chat.ID, fmt.Sprintf("Here is the script of your audio:\n\n%v", text))
	if !isPremium {
		// Add time spent to redis
		timeSpent, err := redisService.IncrementUnpremiumTime(update.Message.From.ID, update.Message.Audio.Duration)
		if err != nil {
			log.Println(err)
		}
		// Send decoded text to user
		remainingTime := unpremiumMaxAudioDuration - timeSpent
		minutes := remainingTime / 60
		seconds := remainingTime % 60
		sendMessage(bot, update.Message.Chat.ID, fmt.Sprintf("Remaining free time: %02d minutes %02d seconds", minutes, seconds))
	}
}

// isPremium checks if user is premium
func isPremium(bot *tgbotapi.BotAPI, userId int, chatID int64, redisService *service.RedisService) bool {
	isPremium, err := redisService.IsPremium(userId)
	if err != nil {
		log.Println(err)
		sendMessage(bot, chatID, "There was an error checking if you are a premium user, please try again later")
		return false
	}
	return isPremium
}

// isAudioDurationAllowed checks if audio duration is allowed for user
func isAudioDurationAllowed(audioDuration int, userId int, userName string, isPremium bool, redisService *service.RedisService) error {
	if isPremium {
		return nil
	}
	if audioDuration > unpremiumMaxAudioDuration {
		return errors.New(fmt.Sprintf("The audio file is too long. Only %v minutes allowed for users without premium", unpremiumMaxAudioDuration/60))
	}
	notPremiumTime, err := redisService.GetUnpremiumTimeSpent(userId)
	if err != nil {
		log.Println(err)
		if err.Error() == "Max time exceeded" {
			return errors.New(fmt.Sprintf("Dear %v, You have exceeded maximum numbers of free decoding of audio,to get premium - type /premium", userName))
		}
		if err.Error() == "User doesn't exist" {
			err := redisService.SaveUnpremiumUser(userId)
			if err != nil {
				log.Println(err)
				return errors.New("There is an error occurred, please try again later")
			}
		}
	}
	// Check if user has enough time
	if audioDuration+notPremiumTime > unpremiumMaxAudioDuration {
		remainingTime := unpremiumMaxAudioDuration - notPremiumTime
		minutes := remainingTime / 60
		seconds := remainingTime % 60
		return errors.New(fmt.Sprintf("Too long audio, you dont have enough free time, remaining time: %02d minutes %02d seconds", minutes, seconds))
	}
	return nil
}

// decodeAudioFile decodes audio file to text
func decodeAudioFile(bot *tgbotapi.BotAPI, fileID string, redisService *service.RedisService) (string, error) {
	// Get audio file
	fileURL, err := uploadUserFileData(bot, fileID)
	if err != nil {
		log.Println(err)
		return "", errors.New("There is an error occurred while decoding the file")
	}
	text, err := decodeFile(fileURL)
	if err != nil {
		log.Println(err)
		return "", errors.New("There is an error occurred while decoding the file")
	}
	return text, nil
}

func handleVoice(bot *tgbotapi.BotAPI, update tgbotapi.Update, voice *tgbotapi.Voice, redisService *service.RedisService) {
	msg := tgbotapi.MessageConfig{}
	msg.ChatID = update.Message.Chat.ID
	// Check if user is premium
	isPremium := isPremium(bot, update.Message.From.ID, update.Message.Chat.ID, redisService)

	// Check if voice duration is allowed
	err := isAudioDurationAllowed(update.Message.Voice.Duration, update.Message.From.ID, update.Message.From.FirstName, isPremium, redisService)
	if err != nil {
		sendMessage(bot, update.Message.Chat.ID, err.Error())
		return
	}

	// Send initial message
	sendMessage(bot, update.Message.Chat.ID, fmt.Sprintf("Decoding will take from 15%% to 30%% of file duration if it is not too short"))

	// Decode voice file
	text, err := decodeAudioFile(bot, update.Message.Voice.FileID, redisService)
	if err != nil {
		log.Println(err)
		sendMessage(bot, update.Message.Chat.ID, "There was an error decoding the file")
		return
	}

	// Send decoded text to user
	sendMessage(bot, update.Message.Chat.ID, fmt.Sprintf("Here is the script of your voice:\n\n%v", text))
	if !isPremium {
		// Add time spent to redis
		timeSpent, err := redisService.IncrementUnpremiumTime(update.Message.From.ID, update.Message.Voice.Duration)
		if err != nil {
			log.Println(err)
		}
		// Send decoded text to user
		remainingTime := unpremiumMaxAudioDuration - timeSpent
		minutes := remainingTime / 60
		seconds := remainingTime % 60
		sendMessage(bot, update.Message.Chat.ID, fmt.Sprintf("Remaining free time: %02d minutes %02d seconds", minutes, seconds))
	}
}

// handleGuide handles the /guide command
func handleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update, redisService *service.RedisService) {
	// Create a message config for the response
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	// Set the message text
	msg.Text = fmt.Sprintf("Hello, %v!\n\nI am an AudTextBot! My main ability is to decode your audio/voice in English into text! If you want to do that, just send me an audio/voice in an allowed format. To see the list of allowed formats, type /list.", update.Message.From.FirstName)

	// Send the message
	if _, err := bot.Send(msg); err != nil {
		log.Println("Error sending message:", err)
	}
}

// handlePremium handles the /premium command
func handlePremium(bot *tgbotapi.BotAPI, update tgbotapi.Update, redisService *service.RedisService) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	isPremium, err := redisService.IsPremium(update.Message.From.ID)
	if err != nil {
		log.Println(err)
		sendMessage(bot, update.Message.Chat.ID, "There was an error checking if you are a premium user, please try again later")
		return
	}
	if isPremium {
		msg.Text = "You are already premium user!"
		bot.Send(msg)
		return
	}
	msg.Text = "Payment method is currently on development stage, please try again later.Or you can contact me @Gopher_UA to get a premium"
	bot.Send(msg)
	return
}

// handleList handles the /list command
func handleList(bot *tgbotapi.BotAPI, update tgbotapi.Update, _ *service.RedisService) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.Text = "Here is the list of allowed formats:\n\nAudio: .mp3, .m4a, .ogg, .wav, .flac, .amr\nVoice: .ogg, .oga, .amr, .wav, .flac"
	bot.Send(msg)
	return
}
