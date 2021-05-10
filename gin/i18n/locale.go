package i18n

import (
	"context"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func CurrentLanguage(ctx context.Context, curLang string) string {
	return Localize(
		ctx,
		&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "CurrentLanguage",
				Description: "显示当前请求语言",
				Other:       "当前语言：{{.curLang}}",
			},
			TemplateData: map[string]interface{}{
				"curLang": curLang,
			},
		},
	)
}

func PersonCats(ctx context.Context, username string, catCount int) string {
	return Localize(
		ctx,
		&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "PersonCats",
				Description: "username 有 n 只猫",
				Other:       "<<.username>> 有 <<.count>> 只猫。",
				LeftDelim:   "<<",
				RightDelim:  ">>",
			},
			TemplateData: map[string]interface{}{
				"username": username,
				"count":    catCount,
			},
			// PluralCount 决定该用什么形式
			PluralCount: catCount,
		},
	)
}
