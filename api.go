package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func Chat(req OpenAIRequest) (string, error) {
	// struct -> json
	b, _ := json.Marshal(req)

	// Get base URL from environment variable, with fallback to default
	baseUrl := os.Getenv("LLM_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://api.deepseek.com/chat/completions"
	}

	// create a client
	client := &http.Client{}
	// create a  http request
	request, err := http.NewRequest(http.MethodPost, baseUrl, bytes.NewReader(b))
	if err != nil {
		log.Println("[command]error occured when create a request:", err)
		return "", err
	}

	api := os.Getenv("API_KEY")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+api)
	request.Header.Set("Accept", "application/json")

	// do request
	res, err := client.Do(request)
	if err != nil {
		log.Println("[command]error occured when get a response:", err)
		return "", err
	}
	defer res.Body.Close()

	// handle response
	if res.StatusCode != http.StatusOK {
		log.Println("[command]API returned status:", res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("[command]error reading response:", err)
		return "", err
	}

	log.Println("Raw response:", string(resBody))

	var openAiResponse = OpenAIResponse{}
	err = json.Unmarshal(resBody, &openAiResponse)
	if err != nil {
		log.Println("[command]error occured when parse a response:", err)
		return "", err
	}

	log.Println("response:", openAiResponse)

	// get the ai response
	msg := openAiResponse.Choices[0].Message.Content
	return msg, nil
}

func removeJsonTag(str string) string {
	s := strings.Replace(str, "```json", "", 1)
	s = strings.Replace(s, "```", "", 1)
	return strings.TrimSpace(s)

}
