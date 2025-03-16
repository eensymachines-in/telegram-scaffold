package watch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/eensymachines-in/telegram-scaffold/models"
	log "github.com/sirupsen/logrus"
)

// WatchUpdates is a function that watches for updates from the bot in an infinite loop
// and returns a channel that can be used to read updates from the bot
// ctx is the context that is used to cancel the function
// bot is the bot instance that is used to fetch updates
// bot fetch interval has to be atleast 5 seconds, typical value is 30 seconds
// errLimit is the number of times the function can fail before it exits the loop, this has to always greater than 0
/*
	func TestGetUpdates(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		bot := models.NewMyBot(123456789, "123456789:ABCDEF", 30*time.Second)
		for updt  := range WatchUpdates(ctx, bot, 5*time.Second, 5) {
			for _, u := range updt {
				if u.Message != nil {
					log.Infof("📩 message received: %s", u.Message.Text)
				} else if u.ChnPost != nil {
					log.Infof("📩 channel post received: %s", u.ChnPost.Text)
				}
			}
		}
*/
func WatchUpdates(ctx context.Context, bot *models.MyBot, handler FuncBotMessageParse, httpTimeOut time.Duration, errLimit uint16) chan []models.BotResponse {
	offset := int64(0)
	tooManyErrors := uint16(0) // count of errors when processing in a loop
	updates := make(chan []models.BotResponse)
	botBaseURL := fmt.Sprintf("%s%s", models.TelegOrgBotBaseURL, bot.Token)
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		Timeout: httpTimeOut,
	} // single http client thats used to fetch updates
	go func() {
		defer close(updates)
		defer log.Warn("⚠ WatchUpdates() now shutting down...")
		log.Infof("🔍 watching updates..")
		for {
			url := ""
			if errLimit == tooManyErrors {
				log.Errorf("❌ too many errors, exiting..")
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(bot.FetchUpdateInterval):
				var err error
				var resp *http.Response
				if offset == 0 {
					url = fmt.Sprintf("%s/getUpdates", botBaseURL)
				} else {
					url = fmt.Sprintf("%s/getUpdates?offset=%d", botBaseURL, offset)
				} //url changes depending on the offset
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					log.Errorf("failed to initiate http request: %s", err)
					return
				}
				resp, err = httpClient.Do(req) // if server unresponsive, headers would not be sent, and this often fails
				if err != nil {
					// context deadline exceeded (Client.Timeout exceeded while awaiting headers)
					// this is a timeout error, we can retry
					// failed to initiate http request: Get "https://api.telegram.org/bot<token>/getUpdates?offset=0": dial tcp: i/o timeout
					// happens after repeated regular attempts to get updates from telegram server.
					// But this does not seem to be a network issue, but a server issue
					// we can wait for a few seconds and try another request
					// if the retry fails, we can continue to the next iteration
					<-time.After(5 * time.Second)
					resp, err = httpClient.Do(req)
					if err != nil {
						log.Errorf("❌ failed cl.Do(): Server unresponsive %s", err)
						continue
					}
				}
				if resp.StatusCode != http.StatusOK {
					log.Errorf("❌ unexpected response code from server %d", resp.StatusCode)
					tooManyErrors++
					continue
				}
				byt, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Errorf("❌ failed to read response from telegram server %s", err)
					tooManyErrors++
					continue
				}
				resp.Body.Close()
				result := models.UpdateResponse{}
				if err := json.Unmarshal(byt, &result); err != nil {
					log.Errorf("❌ failed to unmarshall response from telegram server %s", err)
					tooManyErrors++
					continue
				}
				if len(result.Result) > 0 { // only if there are any updates, if not the update continues to be the previous one
					log.Debug("✅ updates received..")
					offset = result.Result[len(result.Result)-1].Id + 1 // next update is one ahead of the last update offset
					allResponses := []models.BotResponse{}
					for _, updt := range result.Result {
						botRes, err := handler(updt.Message.Text)
						if err != nil {
							log.WithFields(log.Fields{
								"message": updt.Message.Text,
								"err":     err,
							}).Error("❌ failed to parse botmessage")
						} else {
							allResponses = append(allResponses, botRes)
						}
					}
					// Dispatching the parsed responses out on the channel
					if len(allResponses) > 0 {
						updates <- allResponses
					}
				} else {
					continue
				}
			}
		}
	}() // end of go routine
	return updates
}
