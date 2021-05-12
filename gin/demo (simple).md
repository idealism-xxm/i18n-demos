## 简介

在 [Django i18n demo](../django/demo%20(simple).md) 中介绍了国际化和本地化的相关概念，以及如何在 Django 中进行国际化和本地化，主要讲解了实际使用的语言翻译和本地时区两部分。本文继续介绍如何在 Gin 中进行国际化和本地化。

## 语言翻译

Gin 本身不包含国际化和本地化相关的功能，需要通过其他的库实现。我们使用的是 [go-i18n](https://github.com/nicksnyder/go-i18n) 这个库，它不与 Gin 强绑定，可以同时直接用于其他场景（例如：微服务和异步任务等）。

go-i18n 有以下特点：

- 支持 200 多中语言及单复数
  - 复合 [CLDR 复数规则](https://www.unicode.org/cldr/cldr-aux/charts/28/supplemental/language_plural_rules.html)
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

`i18n.Message` 定义了 Zero, One, Two, Few, Many, Other 这些字段，用于表示不同复数下的翻译串。每种语言的复数规则不一样，只需要按照对应语言的复数规则填写对应的字符串即可，没有则不需要填写。

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
# 该文件被删除了，因为唯一的一个消息在最新的参照没有

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

经过翻译后，增量翻译文件如下：

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

不过平时使用时，后续更多得是少量增量的翻译，手动处理更新可能效率更高，或者直接使用翻译平台提供的的集成插件即可。

### 抽取文本


### 编译文本

### 测试

我们在实际场景中通过请求头中的 `Accept-Language` 来传递并获取用户使用的语言。测试代码如下：

```python3
# urls.py
urlpatterns = [
    path('admin/', admin.site.urls),
    re_path(r'hello/(?P<username>\w+)/', views.hello),
]

# constants.py
CURRENT_LANGUAGE = 'Current Language: {cur_lang}'
HELLO = '你好，{username}。'

# views.py
def hello(request, username):
    return HttpResponse(
        f"{_t(CURRENT_LANGUAGE).format(cur_lang=get_language())}<br/>"
        f"{_t(HELLO).format(username=username)}"
    )
```

不同请求头测试结果如下（可以使用 [ModHeader](https://chrome.google.com/webstore/detail/modheader/idgpnmonknjnojddfkpgkljpfnnfcklj) 插件在浏览器中快速修改请求头）：

```
# Accept-Language: zh-CN,zh;q=0.9,en;q=0.8
当前语言: zh-hans
你好，idealism。

# Accept-Language: en-US
Current Language: en-us
Hello, idealism.

# Accept-Language: zh-TW
当前语言: zh-hans
你好，idealism。
```

## 本地时区

时区一般不需要后端做特殊处理，后端统一返回给前端 UTC 时间，让前端自动转换即可。但我们实际使用场景中有导出文件的功能，需要展示用户端本地时区的时间，所以还是需要处理一下本地时区。

### 确定传递形式

本地时区的信息可以和语言翻译保持一致，通过请求头进行传递。而 Django 默认使用时区信息是 TZ database name （例如： `Asia/Shanghai` ），那么我们期望前端也能获得这样的信息，前端刚好有相应的 api 获取：

```javascript
// 获取客户端的时区信息 
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Intl
new Intl.DateTimeFormat().resolvedOptions().timeZone
```

### 中间件中激活本地时区

中间件实现很简单，就是获取请求头中的时区，不存在时使用全局配置中的时区；然后在当前线程中激活，待请求完毕后再清楚时区信息。

```python3
class LocaleTimezoneMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response

    def __call__(self, request):
        # 1. 获取 header 上的时区 ，不存在则使用全局配置中的时区
        try:
            current_timezone = pytz.timezone(request.headers['X-Timezone'])
        except KeyError:
            current_timezone = pytz.timezone(settings.TIME_ZONE)

        # 2. 在当前线程激活对应的时区
        timezone.activate(current_timezone)

        response = self.get_response(request)

        # 3. 清除时区
        timezone.deactivate()
        return response
```

### 配置时区

```python3
# 配置全局默认时区（默认是 UTC）
TIME_ZONE = 'UTC'

MIDDLEWARE = [
    ...
    # 将本地化时区中间件放入靠前的位置
    'django_i18n.middlewares.LocaleTimezoneMiddleware',
    ...
]
```

这样配置后，所有的请求都和语言翻译一样都能识别客户端的本地时区了，我们可以在需要的地方直接使用。

### 使用本地时区

Django 中提供的可以获取当前线程中本地时区的方法，我们就利用它在需要的地方替换时间的时区即可。

```python3
# 获取本次请求的时区
cur_timezone = timezone.get_current_timezone()

# timezone.now() 返回的时间时区是 settings.TIME_ZONE
# .astimezone(cur_timezone) 替换该时间中的时区信息
timezone.now().astimezone(cur_timezone)
```

### 测试

测试代码和前面类似，直接看一下测试结果：

```
# x-timezone: Asia/Shanghai
(Asia/Shanghai) 2021-05-05 21:55:53.823958+08:00

# x-timezone: Africa/Abidjan
(Africa/Abidjan) 2021-05-05 13:56:53.336653+00:00

# x-timezone: Europe/Berlin
(Europe/Berlin) 2021-05-05 15:58:10.404253+02:00
```

## 小结

Django 提供的关于国际化和本地化的组件非常全面，我们只需要灵活运用即可满足大部分的业务场景。 

实际使用过程中还存在跨微服务、跨任务等场景，需要保证全链路国际化。这些原理基本都类似，主要是需要抽出公共组件简化后续的开发，将在后续文章中继续介绍，下次将先介绍如何在 Gin 中实现国际化。
