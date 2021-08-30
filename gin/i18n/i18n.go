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
		bundle.MustLoadMessageFile(fmt.Sprintf("%v/active.%v.toml", "/idealism-xxm/gin/i18n/translation", lang.String()))
	}
	bundleMatcher = language.NewMatcher(bundle.LanguageTags())

	// 3. 初始化默认本地化的 localizer
	defaultLocalizer = i18n.NewLocalizer(bundle, defaultLanguage.String())
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

func WithLanguageAndTag(ctx context.Context, acceptLanguage string) context.Context {
	// 1. 选择最适合的一个语言（方法和 go-i18n 自带第一步一致）
	languageTags, _, _ := language.ParseAcceptLanguage(acceptLanguage)
	supportedLanguages := bundle.LanguageTags()
	_, index, _ := bundleMatcher.Match(languageTags...)
	languageTag := supportedLanguages[index]

	// 2. 新建一个最适配当前请求的本地化器
	localizer := i18n.NewLocalizer(bundle, languageTag.String())

	// 3. 放入 context 中，然后返回
	ctx = WithLanguageTag(ctx, languageTag)
	ctx = WithLocalizer(ctx, localizer)
	return ctx
}
