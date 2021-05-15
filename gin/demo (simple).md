## 简介

在 [Django i18n demo](../django/demo%20(simple).md) 中介绍了国际化和本地化的相关概念，以及如何在 Django 中进行国际化和本地化，主要讲解了实际使用的语言翻译和本地时区两部分。本文继续介绍如何在 Gin 中进行国际化和本地化。

## 语言翻译

Gin 本身不包含国际化和本地化相关的功能，需要通过其他的库实现。我们使用的是 [go-i18n](https://github.com/nicksnyder/go-i18n) 这个库，它不与 Gin 强绑定，可以同时直接用于其他场景（例如：微服务和异步任务等）。

go-i18n 有以下特点：

- 支持 200 多种语言及单复数
  - 符合 [CLDR 复数规则](https://www.unicode.org/cldr/cldr-aux/charts/28/supplemental/language_plural_rules.html)
- 支持带命名变量的字符串，使用 [text/template](http://golang.org/pkg/text/template/) 的语法
- 支持多种形式的翻译文件（JSON, TOML, YAML）

### 标记文本

go-i18n 也支持标记文本并进行抽取，它通过定义一个 `i18n.Message` 实例进行标记，抽取后会生成相应的翻译文件。

```go
var personCats = &i18n.Message{
    // 唯一标识，用于查询某种语言下对应的翻译信息
    ID:          "PersonCats",
    // 该翻译的描述
    Description: "username 有 n 只猫",
    Other:       "<<.username>> 有 <<.count>> 只猫。",
    // 左分隔符
    LeftDelim:   "<<",
    // 右分隔符
    RightDelim:  ">>",
}
```

`i18n.Message` 定义了 Zero, One, Two, Few, Many, Other 这些字段，用于表示不同单复数下的翻译串。每种语言的复数规则不一样，只需要按照对应语言的复数规则填写对应的字符串即可，没有则不需要填写。

- 例如：汉语基本可以只填 Other 一个字段，英语可以根据需要填 One 和 Other 两个字段即可。

LeftDelim 和 RightDelim 字段用于指定分隔符，当我们的文本含有默认分隔符 `{{` 或 `}}` ，可以指定避免错误。这种方式很方便，因为不同语言下的分隔符可能有不同含义，这样新增的语言可以配置独立其他的分隔符。

- 注意：分隔符指定仅对当前 `i18n.Message` 生效，每种语言的每个翻译消息都会生成独立的 `i18n.Message` ，都需要进行单独配置。

### 抽取文本并翻译

当我们定义好时，就可以运行自带的命令行工具抽取文本了，首先需要执行以下命令进行安装：

```shell
go get -u github.com/nicksnyder/go-i18n/v2/goi18n
```

该命令行工具提供了两个子命令可供使用：

- extract: 抽取指定源文件中的文本，并生成某个语言指定格式的翻译文件 `active.*.format`
- merge: 读取所有指定的翻译文件，并对其中每个语言生成两个文件
  - `active.*.format`: 包含运行时需要加载的消息
  - `translate.*.format`: 包含待翻译的消息

第一个子命令很好理解，第二个子命令不太容易明白如何使用，可以通过一个简单的例子来说明如何使用两个子命令快速生成增量的待翻译文件。

新增语言和新增消息其实是一样的，新增语言相当于对原本为空的翻译文件增加了全部消息。所以我们只考虑新增消息这种场景，新增语言时只需要新建一个空的翻译文件即可。

#### 抽取原始数据

我们选择一种实际不会使用到的语言（例如 `zh`）用于抽取翻译文件（如果源代码中所有的消息都是一种语言，当然可以直接指定为它），以便后续增量更新。

```shell
goi18n extract --sourceLanguage zh --outdir i18n/translation
```

生成文件如下：

```toml
[CurrentLanguage]
description = "显示当前请求语言"
other = "当前语言：{{.curLang}}"

[PersonCats]
description = "username 有 n 只猫"
other = "<<.username>> 有 <<.count>> 只猫。"
# 注意：目前不会生成分隔符号，需要翻译人员手动指定
```

#### 生成增量的待翻译文件

假设我们原有的翻译文件如下：

```toml
# i18n/translation/active.en-US.toml
[CurrentLanguage]
description = "显示当前请求语言"
other = "Current Language: <<.curLang>>"
leftDelim = "<<"
rightDelim = ">>"

# i18n/translation/active.zh-Hans.toml
[MyCats]
description = "我有 n 只猫"
other = "我有 {{.count}} 只猫。"
```

我们现在生成 `en-US` 和 `zh-Hans` 语言的增量待翻译文件（如果第一次生成，需要新建一个空白文件），可以直接运行以下命令即可：

```shell
goi18n merge --sourceLanguage zh --outdir i18n/translation i18n/translation/active.zh.toml i18n/translation/active.zh-Hans.toml i18n/translation/active.en-US.toml
```

注意：这个命令，会根据 sourceLanguage 指定的文件作为参照进行更新

- 如果 `active.zh.toml` 删除了某个消息，其他所有 `active.*.toml` 都会直接删除该消息
- 如果 `active.zh.toml` 修改了某个消息的描述，其他所有 `active.*.toml` 都会直接修改该消息的描述，不会修改翻译字符串
- 如果 `active.zh.toml` 新增了某个消息的描述，其他所有 `translate.*.toml` 都会新增该消息，用于增量翻译

生成文件如下：

```toml
# i18n/translation/active.en-US.toml
[CurrentLanguage]
description = "显示当前请求语言"
hash = "sha1-bc7c9ba2dd5c4ea9d428a997a4083055752e723a"
other = "Current Language: <<.curLang>>"
# 注意：目前不会生成分隔符号，需要翻译人员手动指定，所以这里分隔符被删除了

# i18n/translation/translate.en-US.toml
# 新增了 en-US 的增量待翻译文件
[PersonCats]
description = "username 有 n 只猫"
hash = "sha1-5001f112730310aa294aae74838fcf99c9b35467"
other = "<<.username>> 有 <<.count>> 只猫。"

# i18n/translation/active.zh-Hans.toml
# 该文件被删除了，因为唯一的一个消息在最新的参照中没有

# i18n/translation/translate.zh-Hans.toml
[CurrentLanguage]
description = "显示当前请求语言"
hash = "sha1-bc7c9ba2dd5c4ea9d428a997a4083055752e723a"
other = "当前语言：{{.curLang}}"

[PersonCats]
description = "username 有 n 只猫"
hash = "sha1-5001f112730310aa294aae74838fcf99c9b35467"
other = "<<.username>> 有 <<.count>> 只猫。"
```

#### 增量更新翻译文件

经过人工翻译后，增量翻译文件如下：

```toml
# i18n/translation/translate.en-US.toml
[PersonCats]
description = "username有 n 只猫"
one = "<<.username>> have <<.count>> cat."
other = "<<.username>> have <<.count>> cats."
leftDelim = "<<"
rightDelim = ">>"

# i18n/translation/translate.zh-Hans.toml
[CurrentLanguage]
description = "显示当前请求语言"
other = "当前语言：{{.curLang}}"

[PersonCats]
description = "username 有 n 只猫"
other = "{{.username}} 有 {{.count}} 只猫。"
```

此时我们还需要在此运行 merge 子命令，将每个语言的增量翻译合并到 `active.*.toml` 中：

```shell
# 合并增量翻译，并删除增量翻译文件
goi18n merge --sourceLanguage en-US --outdir i18n/translation i18n/translation/active.en-US.toml i18n/translation/translate.en-US.toml && rm i18n/translation/translate.en-US.toml
# 合并增量翻译，并删除增量翻译文件
goi18n merge --sourceLanguage zh-Hans --outdir i18n/translation  i18n/translation/translate.zh-Hans.toml && rm i18n/translation/translate.zh-Hans.toml
```

合并后翻译结果如下：

```toml
# i18n/translation/active.en-US.toml
[CurrentLanguage]
description = "显示当前请求语言"
other = "Current Language: <<.curLang>>"

[PersonCats]
description = "username有 n 只猫"
one = "<<.username>> have <<.count>> cat."
other = "<<.username>> have <<.count>> cats."
# 注意：目前不会生成分隔符号，合并后这里需要再次手动指定

# i18n/translation/active.zh-Hans.toml
[CurrentLanguage]
description = "显示当前请求语言"
other = "当前语言：{{.curLang}}"

[PersonCats]
description = "username 有 n 只猫"
other = "{{.username}} 有 {{.count}} 只猫。"
```

目前该工具有前面提到的小问题，如果使用了自定义分隔符，可能会很容易出错。没有使用自定义分隔符时，则可以极大简化翻译流程的各项操作。

不过平时使用时，后续更多的是少量增量的翻译，手动处理更新可能效率更高，或者直接使用翻译平台提供的的集成插件即可。

### 初始化翻译信息

现在我们已经具备了全部的翻译文件，接下来就是在启动时将这些信息读入到程序中，并配置默认语言和本地化器，方便后续为每个请求执行翻译。

```go
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
```

### 中间件中选择本地化器

我们可以在 Gin 中间件中选择当前请求最合适的语言和本地化器，并将它们都放入到 ctx 中，供后续流程本地化时使用。

这里将语言也需要将语言放入到 ctx 中，因为无法从本地化器中获取对应的语言，而后续在使用微服务、任务等功能时，需要将语言信息传递下去。

```go
// GinMiddleware returns a gin middleware that
// finds a best localizer for current request and adds it to context
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取客户端接受的语言，并选择最适合的一个语言（方法和 go-i18n 自带第一步一致）
		acceptLanguage := c.GetHeader("Accept-Language")
		languageTags, _, _ := language.ParseAcceptLanguage(acceptLanguage)
		supportedLanguages := bundle.LanguageTags()
		_, index, _ := bundleMatcher.Match(languageTags...)
		languageTag := supportedLanguages[index]

		// 2. 新建一个最适配当前请求的本地化器
		localizer := i18n.NewLocalizer(bundle, languageTag.String())

		// 3. 放入 context 中
		ctx := c.Request.Context()
		ctx = WithLanguageTag(ctx, languageTag)
		ctx = WithLocalizer(ctx, localizer)

		// 4. 替换 context 为最新的
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
```

### 进行本地化

此时所需的前置功能准备好了，我们可以直接在需要使用的地方调用对应的函数即可获得翻译好的文本。

```go
// 为了方便和复用，所有需要翻译的文本都仅对外提供函数式的调用方式
ctx := c.Request.Context()
personCat := i18n.PersonCats(ctx, username, 1)
personCats := i18n.PersonCats(ctx, username, 2)
```

翻译函数会在内部处理好所有的翻译相关的逻辑，值得注意的是：除了我们需要传递给翻译模版的全部参数外，有单复数时还需要传递决定单复数形式的参数 PluralCount 。

由于只能有一个 PluralCount 参数，所以当一个需要翻译的文本有多个单复数参数时，最好拆开分成多次处理。不过这样还是会存在某些语言的顺序问题无法处理，但是这样的文本很少出现，最好可以修改文案，从源头拆分。

```go
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
```

### 测试

我们在实际场景中通过请求头中的 `Accept-Language` 来传递并获取用户使用的语言。测试代码如下：

```python3
r.GET("/hello/:username/", func(c *gin.Context) {
    // 获取路径中的 username
    username := c.Param("username")
    // 从 context 中获取 languageTag
    langaugeTag := i18n.LanguageTagFromContext(c.Request.Context())
    currentLanguage := i18n.CurrentLanguage(c.Request.Context(), langaugeTag.String())
    personCat := i18n.PersonCats(c.Request.Context(), username, 1)
    personCats := i18n.PersonCats(c.Request.Context(), username, 2)
    c.String(http.StatusOK, fmt.Sprintf("%v\n%v\n%v", currentLanguage, personCat, personCats))
})
```

不同请求头测试结果如下（可以使用 [ModHeader](https://chrome.google.com/webstore/detail/modheader/idgpnmonknjnojddfkpgkljpfnnfcklj) 插件在浏览器中快速修改请求头）：

```
# Accept-Language: zh-CN,zh;q=0.9,en;q=0.8
当前语言：zh-Hans
idealism 有 1 只猫。
idealism 有 2 只猫。

# Accept-Language: en
Current Language: en-US
idealism have 1 cat.
idealism have 2 cats.

# Accept-Language: zh-TW
当前语言：zh-Hans
idealism 有 1 只猫。
idealism 有 2 只猫。
```

## 本地时区

正如 [Django i18n demo](../django/demo%20(simple).md) 提到的，时区一般不需要后端做特殊处理，后端统一返回给前端 UTC 时间，让前端自动转换即可。但我们实际使用场景中有导出文件的功能，需要展示用户端本地时区的时间，所以还是需要处理一下本地时区。

### 确定传递形式

Go 中时区也是通过 TZ database name （例如： `Asia/Shanghai` ）获取时区信息。我们已经在上一篇文章中介绍过了，前端可以通过以下方法直接获取到相应的时区信息：

```javascript
// 获取客户端的时区信息 
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Intl
new Intl.DateTimeFormat().resolvedOptions().timeZone
```

### 中间件中激活本地时区

中间件实现很简单，就是获取请求头中的时区，不存在时使用全局配置中的时区，然后在当前 ctx 中激活即可。

```go
// TzMiddleware returns a gin middleware that
// finds a best `*time.Location` for current request and adds it to context
func TzMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 获取客户端指定的时区
        tzName := c.GetHeader("X-Timezone")
        location, err := time.LoadLocation(tzName)
        if tzName == "" || err != nil {
        // 没有指定时区 或者 出错，则使用默认时区
            location = defaultLocation
        }
        
        // 2. 放入 context 中
        ctx := WithLocation(c.Request.Context(), location)
        
        // 3. 替换 context 为最新的
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
  }
```

### 使用本地时区

我们直接调用提供的函数从 ctx 中获取对应的时区实例即可：

```go
location := LocationFromContext(ctx)
nowStr := time.Now().In(location).Format(time.RFC3339Nano)
timeStr := fmt.Sprintf("(%s) %s", location.String(), nowStr)
```

### 复制时区文件

为了减少镜像体积，我们一般会将编译好的 Go 程序放入到 apline 镜像内部，但它未提供时区文件，这时就需要在打包时同时拷贝时区文件，否则无法成功运行需要使用时区的程序。

```dockerfile
FROM golang:1.15.4 as builder

WORKDIR /idealism-xxm/gin
ADD . .

RUN CGO_ENABLED=0 GOOS=linux go build -v -ldflags "-s -w" -installsuffix cgo -o . ./...

FROM alpine:latest
WORKDIR /idealism-xxm/gin
# alpine 没有提供时区信息，需要从 builder 拷过来
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /opt/zoneinfo.zip
ENV ZONEINFO /opt/zoneinfo.zip

# 拷贝编译后的程序
COPY --from=builder /idealism-xxm/gin/ /idealism-xxm/gin/
# 添加 i18n 翻译文件
ADD i18n/translation /idealism-xxm/gin/i18n/translation

ENTRYPOINT ["/idealism-xxm/gin/gin-i18n"]
CMD ["run", "--logtostderr"]
```

### 测试

不指定时区和指定两个时区的测试结果如下：

```
# x-timezone: 
(Asia/Shanghai) 2021-05-14T21:10:36.818908+08:00

# x-timezone: Africa/Abidjan
(Africa/Abidjan) 2021-05-14T21:12:57.396036Z

# x-timezone: Europe/Berlin
(Europe/Berlin) 2021-05-14T21:13:35.754335+02:00
```

## 小结

Gin 虽然不提供关于国际化和本地化的组件，但使用 go-i18n 和标准库就足以支持语言翻译和本地时区，而且不与 Gin 框架绑定，可以直接用于微服务、任务等其他景。 

[Django i18n demo](../django/demo%20(simple).md) 和本文分别介绍了如何在 Django 和 Gin 中分别实现独立的国际化和本地化，实际场景中还存在跨服务、跨任务的国际化需求，需要保证全链路国际化。

这种跨服务同步信息的本质就是传递额外信息（例如： traceId, 语言和时区等），下次将先介绍如何在使用 gRPC 的 Python 和 Golang 间实现常见的全链路追踪。
