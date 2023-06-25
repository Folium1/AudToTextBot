package audTextBot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// isPremium checks if user is premium
func isPremium(bot *tgbotapi.BotAPI, userId int, chatID int64) bool {
	isPremium, err := redisService.IsPremium(userId)
	if err != nil {
		log.Println(err)
		sendMessage(bot, chatID, "There was an error checking if you are a premium user, please try again later")
		return false
	}
	return isPremium
}

func checkStatus(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userId := update.Message.From.ID
	premium := isPremium(bot, userId, update.Message.Chat.ID)
	if premium {
		sendMessage(bot, update.Message.Chat.ID, "You are already premium user")
	} else {
		sendMessage(bot, update.Message.Chat.ID, "You are NOT premium user")
	}
}

// handlePremium handles the /premium command
func handlePremium(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
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
	msg.Text = "Payment method is currently on development stage, please try again later.Or you can contact me(@Gopher_UA) to get a premium"
	bot.Send(msg)
	return
}
