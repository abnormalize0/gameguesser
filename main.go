package main

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"log"
	"fmt"
	"bytes"
	"strconv"
	"math/rand"
	"time"
)

func main() {
	tg_token := "5100120720:AAGS9DC56gbm7nYK106hhYC3fnWVzxpqDkY"
	rawg_token := "9f9767607d6b48d1be136afa8bcd6709"
	tg_url := "https://api.telegram.org/bot" + tg_token
	rawg_url := "https://api.rawg.io/api/games?key=" + rawg_token
	offset := 0
	state := 0
	for ;; {
		updates, err := get_updates(tg_url, offset)
		if err != nil {
			log.Println("err: ", err.Error())
		}
		for _, update := range updates {
			state = process(state, tg_url, rawg_url, update)
			//respond(tg_url, update)
			offset = update.UpdateID + 1
		}
		fmt.Println(updates)
	}
}

func process(state int, tg_url string, rawg_url string, update Update) (int) {
	if state == 0 && update.Message.Text == "/start" {
		update.Message.Text = "бла бла правила\n\n1: 70-100\n2:0-100"
		state = 1
	}
	if state == 1 && update.Message.Text == "1" {
		var message BotMessage
		message.ChatID = update.Message.Chat.ChatID
		games := get_random_game(rawg_url, "0,50")
		message.Photo = games.Image[rand.Intn(len(games.Image))].Link
		message.Caption = games.Name
		buf, _ := json.Marshal(message)
		_, _ = http.Post(tg_url + "/sendPhoto", "application/json", bytes.NewBuffer(buf))
	}
	respond(tg_url, update)
	return state
}

func get_updates(tg_url string, offset int) ([]Update, error) {
	resp, err := http.Get(tg_url + "/getUpdates" + "?offset=" + strconv.Itoa(offset))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var tg_response TelegramResponse
	err = json.Unmarshal(body, &tg_response)
	if err != nil {
		return nil, err
	}
	return tg_response.Result, nil
}

func respond(tg_url string, update Update) (error) {
	var message BotMessage
	message.ChatID = update.Message.Chat.ChatID
	message.Text = update.Message.Text
	buf, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = http.Post(tg_url + "/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	return nil
}

func get_random_game(rawg_url string, score string) (RawgUpdate) {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	resp, _ := http.Get(rawg_url + "&metacritic=" + score + "&page_size=1&page=" + strconv.Itoa(r1.Intn(1000)))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var rawg_response RawgResponse
	_ = json.Unmarshal(body, &rawg_response)
	return rawg_response.Result[0]
}