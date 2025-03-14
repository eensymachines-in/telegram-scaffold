package models

type NineDispatchCategory int
type NineMediaType int

const (
	BEEFY NineDispatchCategory = iota
	MEATY
	VEGGIE
)
const (
	JPEG NineMediaType = iota
	MP4
	GIF
)

// BotResponse : interface for mass forwarding messages
// any struct that implements this interface can be used to forward messages
type BotResponse interface {
	TxtMsg() string
	PhotoUrl() string
	VideoUrl() string
	AnimationUrl() string
	MsgCaption() string
	GetChatID() int64 // gets chatid fo  the intended recepient of the bot response
	SetChatID(int64) BotResponse
}

type CategorisedBotResponse interface {
	BotResponse
	GetCategory() string
}

type MultiMediaBotResponse interface {
	BotResponse
	GetMediaType() NineMediaType
}

type NinegagFwdMsg struct {
	Category  string
	MediaTyp  NineMediaType
	Txt       string `json:"text"`
	Photo     string `json:"photo"`
	Video     string `json:"video"`
	Animation string `json:"animation"`
	Caption   string `json:"caption"`
	ChatID    int64  `json:"chat_id"`
}

func (nfm *NinegagFwdMsg) TxtMsg() string {
	return nfm.Txt
}

func (nfm *NinegagFwdMsg) PhotoUrl() string {
	return nfm.Photo
}
func (nfm *NinegagFwdMsg) VideoUrl() string {
	return nfm.Video
}
func (nfm *NinegagFwdMsg) AnimationUrl() string {
	return nfm.Animation
}

func (nfm *NinegagFwdMsg) MsgCaption() string {
	return nfm.Caption
}
func (nfm *NinegagFwdMsg) SetChatID(c int64) BotResponse {
	nfm.ChatID = c
	return nfm
}
func (nfm *NinegagFwdMsg) GetChatID() int64 {
	return nfm.ChatID
}
func (nfm *NinegagFwdMsg) GetCategory() string {
	return nfm.Category
}
func (nfm *NinegagFwdMsg) GetMediaType() NineMediaType {
	return nfm.MediaTyp
}
