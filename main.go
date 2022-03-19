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
	"strings"
)

func main() {
	tg_token := "5100120720:AAGS9DC56gbm7nYK106hhYC3fnWVzxpqDkY"
	rawg_token := "9f9767607d6b48d1be136afa8bcd6709"
	tg_url := "https://api.telegram.org/bot" + tg_token
	rawg_url := "https://api.rawg.io/api/games?key=" + rawg_token
	offset := 0
	state := -1
	for ;; {
		updates, err := get_updates(tg_url, offset)
		if err != nil {
			log.Println("err: ", err.Error())
		}
		for _, update := range updates {
			if state > 0 && strconv.Itoa(state) == update.Message.Text[0:1] {
				update.Message.Text = "Верно!"
				respond(tg_url, update)
				state = 0
			} else if state > 0 {
				update.Message.Text = "Неверно!"
				respond(tg_url, update)
				state = 0
			}
			
			state = process(state, tg_url, rawg_url, update)
			//respond(tg_url, update)
			offset = update.UpdateID + 1
		}
		fmt.Println(updates)
	}
}

func process(state int, tg_url string, rawg_url string, update Update) (int) {
	if state == -1 && update.Message.Text == "/start" {
		update.Message.Text = "Введите 1 чтобы начать."
		state = 0
		respond(tg_url, update)
	} else if state == 0 {
		var message BotMessage
		message.ChatID = update.Message.Chat.ChatID
		games := get_random_game(rawg_url, "0,40")
		message.Photo = games.Image[rand.Intn(len(games.Image))].Link
		//message.Caption = games.Name
		buf, _ := json.Marshal(message)
		_, _ = http.Post(tg_url + "/sendPhoto", "application/json", bytes.NewBuffer(buf))
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		var options [4]string
		answer := r1.Intn(4)
		options[answer] = games.Name
		similar_games := get_similar_game(rawg_url, games.Name)
		counter := 0
		for i := 0; i <4 ; i++ {
			if similar_games[counter] == games.Name {
				counter++
			}
			if options[i] == "" {
				options[i] = similar_games[counter]
				counter++
			}
		}
		_, _ = http.Get(tg_url + "/sendMessage" + "?text=Выберите правильный вариант&chat_id=" + strconv.Itoa(message.ChatID) + "&reply_markup={\"keyboard\":[[\"1: " + strings.ReplaceAll(options[0], "&", "") + "\"],[\"2: " + strings.ReplaceAll(options[1], "&", "") + "\"],[\"3: " + strings.ReplaceAll(options[2], "&", "") + "\"],[\"4: " + strings.ReplaceAll(options[3], "&", "") + "\"]],\"one_time_keyboard\":true,\"resize_keyboard\":true}")
		state = answer + 1
		//respond(tg_url, update)
	}
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
	var rawg_response RawgResponse
	resp, _ := http.Get(rawg_url + "&exclude_additions=true&metacritic=" + score + "&page_size=1&page=" + strconv.Itoa(r1.Intn(1000)))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &rawg_response)
	for (len(rawg_response.Result[0].Tags) < 4) {
		resp, _ := http.Get(rawg_url + "&exclude_additions=true&metacritic=" + score + "&page_size=1&page=" + strconv.Itoa(r1.Intn(1000)))
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &rawg_response)
	} 

	return rawg_response.Result[0]
}
/*
func get_similar_game(rawg_url string, tags []Tags) ([4]string) {
	var similar_games [4]string
	maxtag := len(tags)
	request := "&page=2&tags=" + tags[maxtag - 1].Tag + "&tags=" + tags[maxtag - 2].Tag + "&tags=" + tags[maxtag - 3].Tag + "&tags=" + tags[maxtag - 4].Tag + "&tags=" + tags[maxtag - 5].Tag + "&tags=" + tags[maxtag - 6].Tag + "&tags=" + tags[maxtag - 7].Tag + "&tags=" + tags[maxtag - 8].Tag + "&tags=" + tags[maxtag - 9].Tag + "&tags=" + tags[maxtag - 10].Tag + "&tags=" + tags[maxtag - 11].Tag + "&tags=" + tags[maxtag - 12].Tag
	var rawg_response RawgResponse
	resp, _ := http.Get(rawg_url + request)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &rawg_response)
	for i := 0; i < 4; i++ {
		similar_games[i] = rawg_response.Result[i].Name
	}
	return similar_games
}*/

func get_similar_game(rawg_url string, name string) ([6]string) {
	var similar_games [6]string
	//maxtag := len(tags)
	request := "&exclude_additions=true&search=" + name
	var rawg_response RawgResponse
	resp, _ := http.Get(rawg_url + request)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &rawg_response)
	fmt.Println(len(rawg_response.Result))
	if len(rawg_response.Result) < 6 {
		resp, _ = http.Get(rawg_url)
		defer resp.Body.Close()
		body, _ = ioutil.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &rawg_response)
	} 
	for i := 0; i < 6; i++ {
		similar_games[i] = rawg_response.Result[i].Name
	}
	return similar_games
}