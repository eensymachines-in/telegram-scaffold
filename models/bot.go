package models

/* Typical custom data types for the telegram interaction
// Confiuration, incoming object payloads , outgoing object payloads
// and the bot object itself
*/
import "time"

const (
	TelegOrgBotBaseURL = "https://api.telegram.org/bot" // base URL for the telegram bot API
)

var (
	// NewMyBot : function to create a new instance of the bot
	NewMyBot = func(id int64, tok string, interval time.Duration) *MyBot {
		return &MyBot{
			Token:               tok,
			Id:                  id,
			FetchUpdateInterval: interval,
		}
	}
)

type MyBot struct {
	Token               string        // Unique secret token for the bot
	Id                  int64         // Bot chat ID
	FetchUpdateInterval time.Duration // interval for fetching updates
}

type UpdateResponse struct {
	Ok     bool         `json:"ok"`
	Result []*BotUpdate `json:"result"`
}
type BotUpdate struct {
	Id      int64        `json:"update_id"`
	Message *BotMessage  `json:"message"`
	ChnPost *ChannelPost `json:"channel_post"`
}
type BotMessage struct {
	Id   int64     `json:"message_id"`
	From *FromInfo `json:"from"`
	Chat *ChatInfo `json:"chat"`
	Text string    `json:"text"`
}

type ChannelPost struct {
	Id   int64     `json:"message_id"`
	Chat *ChatInfo `json:"chat"`
	Text string    `json:"text"`
}

type FromInfo struct {
	ID       int64  `json:"id"`
	Fname    string `json:"first_name"`
	Lname    string `json:"last_name"`
	Username string `json:"username"`
}

type ChatInfo struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}
