# Telegram AudToTextBot

This Telegram bot is designed to convert audio/voice messages into text. Simply send an audio/voice message to the bot, and it will process the audio and return the corresponding transcript.

## Prerequisites

Before starting the bot, make sure you have the following:

- Golang (version 1.20.5)
- An API token for your Telegram bot. If you don't have one, you can obtain it by creating a new bot using the Telegram BotFather. Refer to the official Telegram documentation on how to create a bot: [https://core.telegram.org/bots#3-how-do-i-create-a-bot](https://core.telegram.org/bots#3-how-do-i-create-a-bot)
- AssemblyAI API key for getting scripts from audios. You can get it here: [https://www.assemblyai.com](https://www.assemblyai.com)
- Redis server

## How to Start

1. Fill in the environment variables in the `.env` file.
2. Run the following commands in the terminal:
```
make build
make run
```