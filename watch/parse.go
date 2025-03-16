package watch

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/eensymachines-in/telegram-scaffold/models"
)

// parsers are functions that can be extended on this delegate.
// Pass a parser to WatchUpdates to generically handle when the updates are received
type FuncBotMessageParse func(string) (models.BotResponse, error)

var (
	// urlFromItemId : function to generate the url from the item id
	// any departure from the web url, it can be unified to make it look similar to weburl
	urlFromItemId = func(id string) string {
		return fmt.Sprintf("https://img-9gag-fun.9cache.com/photo/%s_460", id)
	}
	// 9gag android app url pattern
	// this is when the user triggers the bot to distribute via the 9gag app
	mobTxtPttrn = regexp.MustCompile(`^https:\/\/9gag.com\/gag\/(?P<itemid>([a-zA-Z0-9]*))\?utm_source=copy_link&utm_medium=post_share\s(?P<category>(#veggie|#meaty|#beefy))\s?(?P<caption>[\w\W\s\d]+)?$`)
	// 9gag url pattern
	// this is when the user triggers the bot to distribute via the 9gag website
	// \p{So} includes all the emoticons
	// . includes all the characters
	// \s allows spaces
	msgTxtPttrn = regexp.MustCompile(`^(?P<url>https:\/\/img-9gag-fun.9cache.com\/photo\/([a-zA-Z0-9]*)_460)(?P<code>(swp|svvp9|svav1|sv))\.(?P<extn>(webm|mp4|webp))\s(?P<category>(#veggie|#meaty|#beefy))\s?(?P<caption>[\w\W\s\d]+)?$`)
)

// NinegagFwdMsgParser : this is what converts the botupdate to whats compatible with BotResponse
// Break down the message using regex parsing
// then determining the media type from the item id
func NinegagFwdMsgParser(msgTxt string) (models.BotResponse, error) {
	// Parse the 9gag website for the latest posts
	mappedResult := map[string]string{} // mapped result of parsing to the regex
	if msgTxtPttrn.MatchString(string(msgTxt)) {
		matches := msgTxtPttrn.FindStringSubmatch(string(msgTxt))
		if len(matches) != len(msgTxtPttrn.SubexpNames()) {
			// Not all the matches as expected
			return nil, fmt.Errorf("invalid message text")
		}
		for i, name := range msgTxtPttrn.SubexpNames() {
			if name != "" && i != 0 {
				mappedResult[name] = matches[i]
			}
		}
	} else if mobTxtPttrn.MatchString(string(msgTxt)) {
		// This when the mobTxtPttrn matches
		// fwding message is sent to the bot using i 9gag app on android
		matches := mobTxtPttrn.FindStringSubmatch(string(msgTxt))
		if len(matches) != len(mobTxtPttrn.SubexpNames()) {
			// Not all the matches as expected
			return nil, fmt.Errorf("invalid message text")
		}
		for i, name := range mobTxtPttrn.SubexpNames() {
			if name != "" && i != 0 {
				mappedResult[name] = matches[i]
			}
		}
		// while the caption and category are captured here, the url will need rectification and normalization to the web url
		mappedResult["url"] = urlFromItemId(mappedResult["itemid"])
	} else { // none of the regular expressions are matching
		return nil, fmt.Errorf("invalid message text")
	}
	fwdMsg := &models.NinegagFwdMsg{Category: mappedResult["category"], Caption: mappedResult["caption"]}
	// ------------- Determing the url , this is specific to 9gag -------------
	url := fmt.Sprintf("%ssv.mp4", mappedResult["url"])
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to determine type of media, %s", err)
	}
	// 9gag server more than often times out and this timeout setting thus has to be a comfy 5 seconds
	cl := &http.Client{Timeout: 15 * time.Second}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to determine type of media, %s", err)
	}
	// url depends on the media type, determining the media type using trial-error GET requests on the 9gag cache
	if resp.StatusCode == http.StatusOK {
		if resp.ContentLength >= int64(99999) {
			fwdMsg.Video = fmt.Sprintf("%ssv.mp4", mappedResult["url"])
			fwdMsg.MediaTyp = models.MP4
		} else {
			fwdMsg.Animation = fmt.Sprintf("%ssv.gif", mappedResult["url"])
			fwdMsg.MediaTyp = models.GIF
		}
	} else if resp.StatusCode == http.StatusNotFound {
		fwdMsg.Photo = fmt.Sprintf("%ss.jpg", mappedResult["url"])
		fwdMsg.MediaTyp = models.JPEG
	} else {
		// if the media type is not recognised, this shall lead to error
		return nil, fmt.Errorf("media none of the recognised formats: %s", err)
	}
	return fwdMsg, nil
}
