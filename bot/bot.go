package bot

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	service "tgbot/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

const (
	updateTimeout = 60
	audioDuration = 300
)

func StartBot() {
	redisService, err := service.NewRedisService()
	if err != nil {
		log.Fatal(err)
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

	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = updateTimeout

	updates, err := bot.GetUpdatesChan(u)

	// Create a channel to receive updates from the updates channel
	updateChannel := make(chan tgbotapi.Update, runtime.NumCPU())

	// Launch a Goroutine to listen for updates and send them to the update channel
	go func() {
		for update := range updates {
			updateChannel <- update
		}
	}()

	// Start multiple Goroutines to handle updates concurrently
	for i := 0; i < runtime.NumCPU()/2; i++ {
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

	if update.Message == nil {
		return
	}

	switch update.Message.Text {
	case "/guide":
		msg.Text = fmt.Sprintf("Hello, %v!", update.Message.From.FirstName)
		bot.Send(msg)
		msg.Text = "I am a AudTextBot! My main ability is to decode your audio into text! If you want to do that, just send me an audio in allowed format."
		bot.Send(msg)
		msg.Text = "To see allowed formats for audio, type /list"
		bot.Send(msg)
	case "/price":
		msg.Text = "Payment methods are in development. You can contact @Gopher_UA to get a premium"
		bot.Send(msg)
	case "/list":
		msg.Text = "Size: up to 50mb\nSupported file formats:\n.3ga\n.8svx\n.aac\n.ac3\n.aif\n.aiff\n.alac\n.amr\n.ape\n.au\n.dss\n.flac\n.flv\n.m4a\n.m4b\n.m4p\n.m4r\n.mp3\n.mpga\n.ogg, .oga, .mogg\n.opus\n.qcp\n.tta\n.vocn\n.wavn\n.wma\n.wv"
		bot.Send(msg)
	default:
		if update.Message.Audio != nil {
			isPremium, err := redisService.IsPremium(update.Message.From.ID)
			if err != nil {
				log.Println(err)
			}

			if update.Message.Audio.Duration > audioDuration && !isPremium {
				msg.Text = "The audio file is too long. Only 5 minutes allowed for users without premium"
				bot.Send(msg)
				return
			}
			if update.Message.Audio.MimeType != "audio/mpeg" {
				msg.Text = "You have sent a file with a bad format. Only .mp3 format is allowed"
				bot.Send(msg)
				return
			}

			if !isPremium {
				_, err = redisService.GetUnpremiumTriesNum(update.Message.From.ID)
				if err != nil {
					if err.Error() == "Max tries exceeded" {
						msg.Text = fmt.Sprintf("Dear %v, You have exceeded maximum numbers of free decoding of audio,to get premium - type /price", update.Message.From.FirstName)
						bot.Send(msg)
						return
					}
					if err.Error() == "User doesn't exist" {
						err := redisService.SaveUnpremiumUser(update.Message.From.ID)
						if err != nil {
							log.Println(err)
							msg.Text = "There is an error occurred, please retry in a few seconds"
							bot.Send(msg)
							return
						}
					}
				}
			}

			msg.Text = fmt.Sprintf("Decoding will take from 15%% to 30%% of file duration")
			bot.Send(msg)

			var text string
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				fileURL, err := uploadUserFileData(bot, update.Message.Audio.FileID)
				if err != nil {
					log.Println(err)
					msg.Text = "There is an error occurred while decoding the file"
					bot.Send(msg)
					wg.Done()
					return
				}
				text, err = decodeFile(fileURL)
				if err != nil {
					log.Println(err)
					msg.Text = "There is an error occurred while decoding the file"
					bot.Send(msg)
					wg.Done()
					return
				}
				wg.Done()
				if !isPremium {
					currentTriesNum, err := redisService.IncrementUnpremiumTriesNum(update.Message.From.ID)
					if err != nil {
						log.Println("Couldn't increment user's tries,err:", err)
						if err.Error() == "Max tries exceeded" {
							msg.Text = fmt.Sprintf("Dear %s, You have exceeded maximum numbers of free decoding of audio,to get premium type /price", update.Message.From.FirstName)
							bot.Send(msg)
						}
						return
					}
					msg.Text = fmt.Sprintf("Free tries left:%d", 3-currentTriesNum)
					bot.Send(msg)
				}
			}()
			wg.Wait()
			msg.Text = fmt.Sprintf("Here is the script of your audio:\n%v", text)
			bot.Send(msg)
			return
		}
	}
}
