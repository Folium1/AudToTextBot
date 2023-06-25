package audTextBot

import (
	"fmt"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// sendMessage sends a message to a chat
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

// handleGuide handles the /guide command
func handleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Create a message config for the response
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	// Set the message text
	msg.Text = fmt.Sprintf("Hello, %v!\n\nI am an %v! My main ability is to decode your audio/voice in English into text! If you want to do that, just send me an audio/voice in an allowed format. To see the list of allowed formats, type /list.", update.Message.From.FirstName, bot.Self.FirstName)

	// Send the message
	if _, err := bot.Send(msg); err != nil {
		log.Println("Error sending message:", err)
	}
}

// handleList handles the /list command
func handleList(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.Text = "Here is the list of allowed formats:\n\nAudio: .mp3, .m4a, .ogg, .wav, .flac, .amr\nVoice: .ogg, .oga, .amr, .wav, .flac"
	bot.Send(msg)
	return
}

func getPremium(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	ownerChatId, err := strconv.ParseInt(os.Getenv("OWNER_CHAT_ID"), 0, 64)
	if err != nil {
		log.Println("Couldn't get owner's chat id:", err)
	}
	if isPremium(bot, update.Message.From.ID, update.Message.Chat.ID) {
		sendMessage(bot, update.Message.Chat.ID, "You are already a premium user!")
		return
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Yes", "yes"),
			tgbotapi.NewInlineKeyboardButtonData("No", "no"),
		),
	)
	msg := tgbotapi.NewMessage(ownerChatId, fmt.Sprintf("Do you give this user @%v a premium? ID: %v", update.Message.From.UserName, update.Message.From.ID))
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
loop:
	for {
		select {
		case answer := <-UserCallBackCh:
			if answer == "yes" {
				err = redisService.SavePremiumUser(update.Message.From.ID)
				if err != nil {
					sendMessage(bot, update.Message.Chat.ID, "There was an error while saving you as a premium user, please contact @Gopher_UA")
					break loop
				}
				sendMessage(bot, update.Message.Chat.ID, "You are now a premium user!!")
				break loop
			}
			if answer == "no" {
				sendMessage(bot, update.Message.Chat.ID, "Sorry, @Gopher_UA didn't allowed you to get premium")
				break loop
			}
		}
	}
}
