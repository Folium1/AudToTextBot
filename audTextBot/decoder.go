package audTextBot

import (
	"bytes"
	"encoding/json"
	"fmt"

	"io/ioutil"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

var (
	transcript_url = "https://api.assemblyai.com/v2/transcript"
)

func decodeFile(file_url string) (string, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	log.Println("Started to decode")
	transcr := transcribe(file_url)
	filetext := poll(transcr)
	return filetext, nil
}

func poll(id string) string {
	log.Println("Started Poll function")
	asseblyApiKey := os.Getenv("ASSEMBLY_API_KEY")
	pollingUrl := transcript_url + "/" + id

	client := &http.Client{}
	req, _ := http.NewRequest("GET", pollingUrl, nil)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", asseblyApiKey)
	var result map[string]interface{}
	for result["status"] != "completed" {
		res, err := client.Do(req)
		if err != nil {
			log.Fatalln(err)
		}
		defer res.Body.Close()
		json.NewDecoder(res.Body).Decode(&result)
	}
	return fmt.Sprintf("%v", result["text"])

}

func transcribe(file_url string) string {
	var asseblyApiKey = os.Getenv("ASSEMBLY_API_KEY")
	// prepare json data
	values := map[string]string{"audio_url": file_url}
	jsonData, _ := json.Marshal(values)

	// setup HTTP client and set header
	client := &http.Client{}
	req, err := http.NewRequest("POST", transcript_url, bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", asseblyApiKey)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// decode json and store it in a map
	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)

	// print the id of the transcribed audio

	id := fmt.Sprintf("%v", result["id"])
	log.Println("Id:", id)
	return id
}

func uploadUserFileData(bot *tgbotapi.BotAPI, fileID string) (string, error) {
	const UPLOAD_URL = "https://api.assemblyai.com/v2/upload"
	ASSEMBLY_asseblyApiKey := os.Getenv("ASSEMBLY_API_KEY")
	fileConfig := tgbotapi.FileConfig{
		FileID: fileID,
	}

	file, err := bot.GetFile(fileConfig)
	if err != nil {
		return "", err
	}
	fileUrl := file.Link(bot.Token)
	resp, err := http.Get(fileUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fileData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// Setup HTTP client and set header
	client := &http.Client{}
	req, _ := http.NewRequest("POST", UPLOAD_URL, bytes.NewBuffer(fileData))
	req.Header.Set("authorization", ASSEMBLY_asseblyApiKey)
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	// decode json and store it in a map
	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)

	file_url := fmt.Sprintf("%v", result["upload_url"])
	fmt.Println(file_url)
	return file_url, nil
}
