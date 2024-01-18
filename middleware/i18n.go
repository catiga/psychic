package ml

import (
	"eli/config"
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle *i18n.Bundle

func init() {
	// 初始化国际化资源束
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustLoadMessageFile(config.Get().Env.En_i818)
	bundle.MustLoadMessageFile(config.Get().Env.Zh_cn_i818)
	// bundle.MustLoadMessageFile("locales/zh_tw.json")
}

func getTranslatedText(code, lang string) string {

	if len(lang) == 0 {
		lang = "zh_cn"
	}
	// 创建本地化对象
	localizer := i18n.NewLocalizer(bundle, lang)
	// 获取翻译文本
	message, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: code})
	if err != nil {
		// 错误处理
		return ""
	}
	return message
}

// ResponseData 是用于返回的通用结构体
type ResponseData struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Succ(lang string, data interface{}) ResponseData {
	return ResponseData{
		Code:    "200",
		Message: getTranslatedText("200", lang),
		Data:    data,
	}
}

func Fail(lang string, code string) ResponseData {
	return ResponseData{
		Code:    code,
		Message: getTranslatedText(code, lang),
		Data:    nil,
	}
}

func Res(lang string, code string, data interface{}) ResponseData {
	return ResponseData{
		Code:    code,
		Message: getTranslatedText(code, lang),
		Data:    data,
	}
}
