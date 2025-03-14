package send

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eensymachines.in/telegram-scaffold/models"
	log "github.com/sirupsen/logrus"
)

// MassFwdAsReceived is a function that forwards messages as received
// ctx is the context that is used to cancel the function
// bot is the bot instance that is used to fetch updates
// chnDispatch is the channel that is used to read updates from the bot
// httpTimeOut is the timeout for the http client
// TODO: this still does not refer to the database for getting the chat details, needs to
func MassFwdAsReceived(ctx context.Context, bot *models.MyBot, chnDispatch chan []models.BotResponse, chatIDFinder Finder) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		Timeout: 15 * time.Second,
	} // single http client thats used to fetch updates
	botBaseURL := fmt.Sprintf("%s%s", models.TelegOrgBotBaseURL, bot.Token)
	for responses := range chnDispatch {
		for _, rsp := range responses {
			go func() {

				var url = "empty/url/that/has/to/be/overwritten"
				if rsp.TxtMsg() != "" {
					url = fmt.Sprintf("%s/sendMessage", botBaseURL)
				} else if rsp.PhotoUrl() != "" {
					url = fmt.Sprintf("%s/sendPhoto", botBaseURL)
				} else if rsp.VideoUrl() != "" {
					url = fmt.Sprintf("%s/sendVideo", botBaseURL)
				} else if rsp.AnimationUrl() != "" {
					url = fmt.Sprintf("%s/sendAnimation", botBaseURL)
				} else {
					log.Error("❌ MassFwdAsReceived: Unrecognized message type")
					return
				}
				chatids, err := chatIDFinder.FindDistributionIds(rsp)
				if err != nil {
					log.WithFields(log.Fields{
						"err": err,
					}).Error("❌ MassFwdAsReceived: unable to get suitable chatids")
					return
				}
				for _, chatid := range chatids {
					byt, err := json.Marshal(rsp.SetChatID(chatid))
					if err != nil {
						log.Error("❌ failed to marshal message")
						return
					}
					req, err := http.NewRequest("POST", url, bytes.NewBuffer(byt))
					if err != nil {
						log.Error("❌ MassFwdAsReceived: failed to create request")
						return
					}
					req.Header.Set("Content-Type", "application/json")
					resp, err := httpClient.Do(req)
					if err != nil || resp.StatusCode != http.StatusOK {
						log.Errorf("❌ MassFwdAsReceived: failed to dispatch message %d : %s", resp.StatusCode, err)
						return
					}
					log.WithFields(log.Fields{
						"caption": rsp.MsgCaption(),
					}).Info("✅ message dispatched")
				}

				return
			}()
		}
	}
}
