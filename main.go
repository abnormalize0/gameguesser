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
	"os"
)

func main() {
	tg_token := "5100120720:AAGS9DC56gbm7nYK106hhYC3fnWVzxpqDkY"
	rawg_token := "9f9767607d6b48d1be136afa8bcd6709"
	tg_url := "https://api.telegram.org/bot" + tg_token
	rawg_url := "https://api.rawg.io/api/games?key=" + rawg_token
	offset := 0
	var states map[int]*Player
	states = make(map[int]*Player)
	lowest := lowest_record()
	for ;; {
		updates, err := get_updates(tg_url, offset)
		if err != nil {
			log.Println("err: ", err.Error())
		}
		for _, update := range updates {
			chatID := update.Message.Chat.ChatID
			if _, ok := states[chatID]; !ok {
				new_challenger := &Player{State: -1, HP: 3, Score: 0}
				states[chatID] = new_challenger
			}
			if states[chatID].HP == 0 {
				if lowest < states[chatID].Score {
					states[chatID].State = -3
				} else {
					states[chatID].State = -4
				}
				
				states[chatID].HP = 3
			}else if states[chatID].State > 0 && strconv.Itoa(states[chatID].State) == update.Message.Text[0:1] {
				states[chatID].Score++
				update.Message.Text = "Верно!"
				respond(tg_url, update)
				states[chatID].State = -2
			} else if states[chatID].State > 0 {
				states[chatID].HP--
				update.Message.Text = "Неверно! Осталось " + strconv.Itoa(states[chatID].HP) + " HP."
				respond(tg_url, update)
				states[chatID].State = -2
			} else if states[chatID].State == -3 {
				save_result(update.Message.Text, states[chatID].Score)
			}
			
			states[chatID].State = process(states[chatID].State, tg_url, rawg_url, update)
			//respond(tg_url, update)
			offset = update.UpdateID + 1
		}
		fmt.Println(updates)
	}
}

func process(state int, tg_url string, rawg_url string, update Update) (int) {
	if state == -1 && update.Message.Text == "/start" {
		update.Message.Text = "Введите 1 чтобы начать. Введите 2 чтобы открыть список лидеров."
		state = -2
		respond(tg_url, update)
	} else if state == -2 && update.Message.Text == "1" {
		var message BotMessage
		message.ChatID = update.Message.Chat.ChatID
		games := get_random_game(rawg_url, "0,40")
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		message.Photo = games.Image[r1.Intn(len(games.Image))].Link
		//message.Caption = games.Name
		buf, _ := json.Marshal(message)
		_, _ = http.Post(tg_url + "/sendPhoto", "application/json", bytes.NewBuffer(buf))
		var options [4]string
		answer := r1.Intn(4)
		options[answer] = games.Name
		similar_games := get_similar_game(rawg_url, games.Tags)
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
	} else if state == -2 && update.Message.Text == "2" {
		leaders := display_records()
		update.Message.Text = leaders
		respond(tg_url, update)
	} else if state == -3 {
		update.Message.Text = "Игра окончена. Установлен рекорд! Введите свой ник:"
		respond(tg_url, update)
	} else if state == -4 {
		update.Message.Text = "че"
		respond(tg_url, update)
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

func get_similar_game(rawg_url string, tags []Tags) ([6]string) {
	var similar_games [6]string
	request_tags := critical_tags(tags)
	request := "&exclude_additions=true" + request_tags
	var rawg_response RawgResponse
	resp, _ := http.Get(rawg_url + request)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &rawg_response)
	pages := rawg_response.Pages
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	//fmt.Println(request)
	similar_games[0] = rawg_response.Result[r1.Intn(20)].Name
	for i := 1; i < 6; i++ {
		request = "&page_size=1&page=" + strconv.Itoa(r1.Intn(pages - 1)) + "&exclude_additions=true" + request_tags
		fmt.Println(request)
		resp, _ = http.Get(rawg_url + request)
		defer resp.Body.Close()
		body, _ = ioutil.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &rawg_response)
		similar_games[i] = rawg_response.Result[0].Name
		
	}
	return similar_games
}

func critical_tags(tags []Tags) (string) {
	var request string
	critical := []string {"open-world", "first-person", "third-person", "Sci-fi", "2d", "horror", "fantasy", "gore", "sandbox", "survival", "exploration", "comedy", "stealth", 
						  "tactical", "action-rpg", "pixel-graphics", "space", "zombies", "anime", "hack-and-slash", "turn-based", "post-apocalyptic", "survival-horror",
						  "cute", "mystery", "side-scroller", "physics", "futuristic", "isometric", "walking-simulator", "roguelike", "parkour", "building", "top-down",
						  "metroidvania", "mmo", "driving", "management", "visual-novel", "puzzle-platformer", "surreal", "3d-platformer", "war", "violent", "dark", "story",
						  "vid-sboku", "platformer-2"}
	for _, tag := range tags {
		for _, critical_tag := range critical {
			//fmt.Println(tag.Tag)
			//fmt.Println(critical_tag)
			if tag.Tag == critical_tag {
				request = request + "&tags=" + critical_tag
			}
		}
	}
	return request
}

func lowest_record() (int) {
	current_min := 999999
	file, _ := os.Open("records.json")
	out, _ := ioutil.ReadAll(file)
	var records []Record
	json.Unmarshal(out, &records)
	for _, record := range records {
		if record.Result < current_min {
			current_min = record.Result
		}
	}
	defer file.Close()
	return current_min
}

func save_result (name string, score int) {
	lowest := lowest_record()
	file, _ := os.Open("records.json")
	out, _ := ioutil.ReadAll(file)
	var records []Record
	json.Unmarshal(out, &records)
	candite := 0
	for i := 0; i < len(records); i++ {
		if records[i].Result == lowest {
			candite = i
			break
		}
	}
	records[candite].Result = score
	records[candite].Name = name
	defer file.Close()
	new_file, _ := json.MarshalIndent(records, "", " ")
	_ = ioutil.WriteFile("records.json", new_file, 0644)

}

func display_records () (string) {
	file, _ := os.Open("records.json")
	out, _ := ioutil.ReadAll(file)
	var records []Record
	json.Unmarshal(out, &records)
	output := "Лидеры:"
	for _, record := range records {
		output = output + "\n" + strconv.Itoa(record.Result) + " " + record.Name
	}
	return output
}