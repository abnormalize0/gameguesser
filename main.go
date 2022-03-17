package main

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"log"
	"fmt"
	"bytes"
	"strconv"
)

func main() {
	token := "5100120720:AAGS9DC56gbm7nYK106hhYC3fnWVzxpqDkY"
	bot_url := "https://api.telegram.org/bot" + token
	offset := 0
	for ;; {
		updates, err := get_updates(bot_url, offset)
		if err != nil {
			log.Println("err: ", err.Error())
		}
		for _, update := range updates {
			err = respond(bot_url, update)
			offset = update.UpdateID + 1
		}
		fmt.Println(updates)
	}
}

func get_updates(bot_url string, offset int) ([]Update, error) {
	resp, err := http.Get(bot_url + "/getUpdates" + "?offset=" + strconv.Itoa(offset))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var rest_response RestResponse
	err = json.Unmarshal(body, &rest_response)
	if err != nil {
		return nil, err
	}
	return rest_response.Result, nil
}

func respond(bot_url string, update Update ) (error) {
	var message BotMessage
	message.ChatID = update.Message.Chat.ChatID
	message.Text = update.Message.Text
	buf, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = http.Post(bot_url + "/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	return nil
}