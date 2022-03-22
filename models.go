package main

type Update struct {
	UpdateID int		`json:"update_id"`
	Message Message		`json:"message"`
}

type Message struct {
	Chat Chat			`json:"chat"`
	Text string			`json:"text"`
}

type Chat struct {
	ChatID int			`json:"id"`
}

type TelegramResponse struct {
	Result []Update		`json:"result"`
}

type RawgResponse struct {
	Result []RawgUpdate	`json:"results"`
	Pages int 			`json:"count"`
}

type BotMessage struct {
	ChatID int			`json:"chat_id"`
	Text string			`json:"text"`
	Photo string		`json:"photo"`
	Caption string		`json:"caption"`
}

type RawgUpdate struct {
	Name string			`json:"name"`
	Image []ScrSht		`json:"short_screenshots"`
	Tags []Tags			`json:"tags"`	
}

type ScrSht struct {
	Link string			`json:"image"`
}

type Tags struct {
	Tag string			`json:"slug"`
}

type Player struct {
	State int
	HP int
}