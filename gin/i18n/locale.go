package i18n

import (
	"context"
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
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

func MyCats(ctx context.Context, catCount int) string {
	func() {
		bundle := i18n.NewBundle(language.AmericanEnglish)
		localizer := i18n.NewLocalizer(bundle, language.AmericanEnglish.String())
		helloPersonMessage := &i18n.Message{
			ID:          "MyCats",
			Description: "我有 n 只猫",
			Other:       "我有 <<.count>> 只猫。",
			LeftDelim:   "<<",
			RightDelim:  ">>",
		}
		fmt.Println(localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: helloPersonMessage,
			TemplateData:   map[string]string{"count": "Nick"},
			PluralCount:    2,
		}))
		// Output:
		// Hello Nick!
	}()
	return Localize(
		ctx,
		&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "MyCats",
				Description: "我有 n 只猫",
				Other:       "我有 <<.count>> 只猫。",
				LeftDelim:   "<<",
				RightDelim:  ">>",
			},
			TemplateData: map[string]interface{}{
				"count": catCount,
			},
			// PluralCount 决定该用什么形式
			PluralCount: catCount,
		},
	)
}
