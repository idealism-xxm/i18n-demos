package i18n

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// 默认语言
var defaultLanguage = language.AmericanEnglish

// 默认本地化器
var defaultLocalizer *i18n.Localizer

// 语言包相关信息
var bundle *i18n.Bundle

// 语言包对应的语言选择器
// 调用 bundle.AddMessage 时，增加的新语言会重新生成 matcher ，
// 我们使用场景只会在启动时加载，所以可以提前获取
var bundleMatcher language.Matcher

func init() {
	// 支持的语言列表
	var languages = []language.Tag{
		language.AmericanEnglish,
		language.SimplifiedChinese,
	}

	// 1. 创建语言包
	bundle = i18n.NewBundle(defaultLanguage)

	// 2. 加载语言文件
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	for _, lang := range languages {
		bundle.MustLoadMessageFile(fmt.Sprintf("%v/active.%v.toml", "/Users/idealism/Workspaces/Go/i18n-in-django-and-gin/gin/i18n/translation", lang.String()))
	}
	bundleMatcher = language.NewMatcher(bundle.LanguageTags())

	// 3. 初始化默认本地化的 localizer
	defaultLocalizer = i18n.NewLocalizer(bundle, defaultLanguage.String())
}

func GetLanguage(ctx context.Context) {
	// 1. 从 context 中获取 localizer
	//localizer := GetLocalizer(ctx)
	//
	//localizer.
}

func Localize(ctx context.Context, lc *i18n.LocalizeConfig) string {
	// 1. 从 context 中获取 localizer
	localizer := LocalizerFromContext(ctx)

	// 2. 本地化
	result, err := localizer.Localize(lc)
	if err != nil {
		fmt.Printf("#localizer.Localize error, error: %v, lc: %v\n", err, lc)
	}
	return result
}
