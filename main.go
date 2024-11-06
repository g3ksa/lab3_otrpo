package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	accessToken   string
	defaultUserID string
	logger        = log.New(os.Stdout, "lab3_otrpo: ", log.LstdFlags)
	fileToSave    = "vk_data.json"
)

type VKData struct {
	UserInfo      interface{} `json:"user_info"`
	Subscriptions interface{} `json:"subscriptions"`
	Followers     interface{} `json:"followers"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	accessToken = os.Getenv("ACCESS_TOKEN")
	defaultUserID = "337773226"
}

func apiRequest(method string, params map[string]string) (interface{}, error) {
	baseURL := "https://api.vk.com/method/"
	params["access_token"] = accessToken
	params["v"] = "5.131"

	url := baseURL + method + "?" + encodeParams(params)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	if errorData, found := result["error"]; found {
		logger.Printf("VK API Error: %v", errorData)
		return nil, fmt.Errorf("VK API Error: %v", errorData)
	}
	response, ok := result["response"]
	if !ok {
		return nil, fmt.Errorf("response key not found in result")
	}

	if responseArray, ok := response.([]interface{}); ok {
		return responseArray, nil
	}
	if responseMap, ok := response.(map[string]interface{}); ok {
		return responseMap, nil
	}

	return nil, fmt.Errorf("unexpected response format")
}

func main() {
	resultFile := flag.String("file_to_save", "", "Куда сохранять данные?")
	userId := flag.String("user_id", "", "По какому пользователю получаем данные?")
	flag.Parse()

	if *userId != "" {
		defaultUserID = *userId
	}
	if *resultFile != "" {
		fileToSave = *resultFile
	}

	getUserMethod := "users.get"
	getUserParams := map[string]string{
		"user_ids": defaultUserID,
		"fields":   "followers_count",
	}
	userInfo, err := apiRequest(getUserMethod, getUserParams)
	if err != nil {
		log.Fatalf("Error fetching user info: %v", err)
	}

	getSubsMethod := "users.getSubscriptions"
	getSubsParams := map[string]string{
		"user_id":  defaultUserID,
		"extended": "1",
	}
	subsInfo, err := apiRequest(getSubsMethod, getSubsParams)
	if err != nil {
		log.Fatalf("Error fetching subscriptions info: %v", err)
	}

	getFollowersMethod := "users.getFollowers"
	getFollowersParams := map[string]string{
		"user_id": defaultUserID,
		"fields":  "first_name,last_name,city,bdate",
	}
	followersInfo, err := apiRequest(getFollowersMethod, getFollowersParams)
	if err != nil {
		log.Fatalf("Error fetching followers info: %v", err)
	}

	vkData := VKData{
		UserInfo:      userInfo,
		Subscriptions: subsInfo,
		Followers:     followersInfo,
	}

	if err := writeToJSONFile(fileToSave, vkData); err != nil {
		log.Fatalf("Error writing data to JSON: %v", err)
	}

	fmt.Println("Данные успешно сохранены в файл vk_data.json.")
}

func writeToJSONFile(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func encodeParams(params map[string]string) string {
	var builder strings.Builder
	for k, v := range params {
		builder.WriteString(fmt.Sprintf("%s=%s&", k, v))
	}
	return builder.String()
}
