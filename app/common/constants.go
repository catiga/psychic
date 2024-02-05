package common

const (
	CODE_SUCCESS                       = 0
	CODE_ERR_METHOD_UNSUPPORT          = 1
	CODE_ERR_REQFORMAT                 = 2
	CODE_ERR_ADMINTOKEN                = 3
	CODE_ERR_NOTFOUND                  = 4
	CODE_ERR_PARAM                     = 5
	CODE_ERR_LAN                       = 901
	CODE_ERR_CHAR_BASPARAM             = 100
	CODE_ERR_CHAR_UNKNOWN              = 101
	CODE_ERR_CHAR_NOTFOUND             = 102
	CODE_ERR_CHAR_ROLE_CAT             = 103
	CODE_ERR_CHAR_PARAM                = 104
	CODE_ERR_CHARBACK_MAX              = 105
	CODE_ERR_CHAR_EXIST                = 106
	CODE_ERR_MISSING_PREREQUISITE_INFO = 107

	CODE_ERR_GPT_COMPLETE   = 201
	CODE_ERR_GPT_STREAM     = 202
	CODE_ERR_GPT_STREAM_EOF = 203
)

const (
	TYPE_CHAT_INITIAL = "chat_init"
	TYPE_CHAT_APPEND  = "chat_follow"
	METHOD_GPT        = "chatGPT"

	CODE_DIRECTION_IN  = "1"
	CODE_DIRECTION_OUT = "2"
)

type Request struct {
	Type      string `json:"type"`
	Method    string `json:"method"`
	Timestamp int64  `json:"timestamp"`
	Ascode    string `json:"ascode"`
	Lan       string `json:"lan"`
	DevId     string `json:"devid"`
	UserId    int64  `json:"userid"`
	Data      string `json:"data"`
	CalId     string `json:"calid"`
	QType     string `json:"qtype"`
}
