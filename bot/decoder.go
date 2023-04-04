package bot

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
	TRANSCRIPT_URL = "https://api.assemblyai.com/v2/transcript"
)

func DecodeFile(file_url string) (string, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	log.Println("Started to decode")
	transcr := Transcribe(file_url)
	filetext := Poll(transcr)
	return filetext, nil
}

func Poll(id string) string {
	log.Println("Started Poll function")
	API_KEY := os.Getenv("ASSEMBLY_API_KEY")
	POLLING_URL := TRANSCRIPT_URL + "/" + id

	client := &http.Client{}
	req, _ := http.NewRequest("GET", POLLING_URL, nil)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", API_KEY)
	var result map[string]interface{}
	for result["status"] != "completed" {
		if result["status"] == "queued" {

		}
		res, err := client.Do(req)
		if err != nil {
			log.Fatalln(err)
		}
		defer res.Body.Close()
		json.NewDecoder(res.Body).Decode(&result)
	}
	return fmt.Sprintf("%v", result["text"])

}

func Transcribe(file_url string) string {
	var API_KEY = os.Getenv("ASSEMBLY_API_KEY")
	// prepare json data
	values := map[string]string{"audio_url": file_url}
	jsonData, _ := json.Marshal(values)

	// setup HTTP client and set header
	client := &http.Client{}
	req, err := http.NewRequest("POST", TRANSCRIPT_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", API_KEY)
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
	ASSEMBLY_API_KEY := os.Getenv("ASSEMBLY_API_KEY")
	fileConfig := tgbotapi.FileConfig{
		FileID: fileID,
	}

	file, err := bot.GetFile(fileConfig)
	if err != nil {
		return "", err
	}
	url := file.Link(bot.Token)
	resp, err := http.Get(url)
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
	req.Header.Set("authorization", ASSEMBLY_API_KEY)
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
