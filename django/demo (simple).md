## 背景

我们项目需要扩展海外业务，所以需要进行国际化 (i18n) 和本地化 (L10n) ，以便于适应不同市场的语言、地区差异以及技术需要。项目后端有两处入口，分别是 Django 和 Gin ，由于 Django 自带完善的国际化与本地化功能，所以先尝试了在 Django 中实现。

## 简介

国际化和本地化看起来区别不大，但是它们的侧重点不一样，[它们之间是互补的，并且它们合起来才能让一个系统适用于各地](https://zh.wikipedia.org/wiki/%E5%9B%BD%E9%99%85%E5%8C%96%E4%B8%8E%E6%9C%AC%E5%9C%B0%E5%8C%96) ：

- 国际化：英文为 internationalization ，简称 i18n
    - 指在设计软件，将软件与特定语言及地区脱钩的过程。当软件被移植到不同的语言及地区时，软件本身不用做内部工程上的改变或修正
    - 一般由开发者在开发时做一次
- 本地化：英文为 localization ，简称 L10n
    - 指当移植软件时，加上与特定区域设置有关的信息和翻译文件的过程
    - 一般由翻译者针对不同的区域各做一次

用一个语言翻译中的简单例子可以更好地说明它们的差异以及国际化和本地化的合作方式：

```python3
# 国际化前
return f'你好，{username}。'
# 国际化后
ori_text = '你好，{username}。'
trans_text = hello_l10n_dict.get(user_lang, text)
return trans_text.format(username=username)

# 本地化
hello_l10n_dict = {
    'en-us': 'Hello, {username}.',
    'zh-hans': '你好，{username}。',
    ...
}
```

国际化相当于将原本代码中固定的地方写活了，将原本编译时确定的一种固定处理方式，推迟到运行时再确定，使得本地化时可以根据用户的语言等信息，定制化确定该使用何种方式进行处理。

国际化与本地化不仅限于语言翻译，还包括本地时区、文字方向、货币、单复数等。由于我们项目后端涉及返回用户信息的地方只有报错信息和导出文件等，我们实际只使用了语言翻译和本地时区两部分。

## 语言翻译

Django 在语言翻译这部分提供的功能十分完善，自带的组件基本都支持了语言翻译。

### 配置开启本地化

默认情况下 Django 不会打开本地化，需要手动进行开启，修改 `settings.py` 文件中的如下内容（没有则添加）：

```python3
MIDDLEWARE = [
    # 将本地化中间件放入靠前的位置
    'django.middleware.locale.LocaleMiddleware',
    ...
]

TEMPLATES = [
    {
        ...
        'OPTIONS': {
            'context_processors': [
                # 如果使用的模版需要 i18n ，则需要加上该上下处理器
                'django.template.context_processors.i18n',
                ...
            ],
        },
    },
]

# 开启 i18n 和 L10n （默认是开启的）
USE_I18N = True
USE_L10N = True

# 设置默认语言（默认是 en-us ）
LANGUAGE_CODE = 'en-us'

# 添加翻译文件夹路径
PROJECT_PATH = os.path.dirname(os.path.dirname(__file__))
LOCALE_PATHS = (
    os.path.join(PROJECT_PATH, 'locale/'),
)

# 设置支持的语言
LANGUAGES = (('en-us', 'American English'), ('zh-hans', 'Simplified Chinese'))
```

Django 自带的本地化中间件会 [使用按照以下顺序找到本次请求本地化的语言](https://docs.djangoproject.com/en/3.2/topics/i18n/translation/#how-django-discovers-language-preference) ：

- url 中的语言前缀
- cookie 中的 `LANGUAGE_COOKIE_NAME`
- 请求头中的 `Accept-Language`
- 全局配置中的 `LANGUAGE_CODE`

Django 在会按照以下顺序选择翻译后的字符串：

- `LOCALE_PATHS` 中列出的目录（按先后顺序）
- `INSTALLED_APPS` 中每个应用下的 `locale/` 目录（按先后顺序）
- `django/conf/locale` 中 Django 提供的基础翻译
- 原字符串

### 标记文本

此时我们可以标记我们需要翻译的所有文本， Django 提供了以下函数帮助我们进行标记，并会在运行时自动选择对应的翻译：

- gettext/ugettext: 将给定的字符串翻译成选择的语言
    - 使用方式： `gettext(message)`
- ngettext/ungettext: 支持单复数的翻译（每种语言的单复数处理逻辑可能不同，所以不要自己实现，直接使用该函数即可）
    - 使用方式： `ngettext(singular, plural, number)`
- pgettext: 带上下文的翻译（同一个字符串在不同上下文中有不同含义时（例如：`May` ），需要指定上下文）
    - 使用方式：`pgettext(context, message)`
- npgettext: 带上下文且支持单复数的翻译
    - 使用方式：`npgettext(context, singular, plural, number)`

上述每个函数都有对应的惰性函数，命名方式为 `xx_lazy` ，这使得在字符串上下文里使用字符串时，才会进行翻译。

更多使用方式可以查看 [官方文档](https://docs.djangoproject.com/en/3.2/topics/i18n/translation/) 。

在实际使用中，我们结合了以下使用场景和具体情况，按照特定的形式进行处理：

- 这些函数仅用于标记，需要减少对源代码的干扰，所以我们会用别名的方式，例如：`_`, `_t` 等
- 实际场景中， `_` 经常被用于忽略函数返回值，为了保持统一且便于全局搜索出使用翻译的地方，我们最终决定 `_t` 作为别名
- 大部分情况下，我们都不需要处理单复数和上下文，所以我们都使用 `ugettext` 函数。该场景下，我们都抽取了一个常量以尽可能复用字符串，也方便后续快速修改

最终使用的形式如下：

```
from django.utils.translation import ugettext as _t

HELLO = '你好，{username}。'

# 注意这里使用了 format ，所以不能使用惰性函数
_t(HELLO).format(username=username)
```

### 抽取文本

当我们将所有需要翻译的文本用上述方法处理好后，就可以直接抽取文本了。

Django 提供的 `makemessages` 可以自动抽取出我们前面标记的文本。

```shell
# 安装依赖的 gettext
apt-get update && apt-get install -y gettext
# 抽取文本
python manage.py makemessages -l en_us -l zh_hans
```

运行后就会在 `locale/` 下面生成两种语言的 `django.po` 文件了，当翻译者全部翻译完成后，再放回对应的位置即可。

文件内容大致如下：

```
...
msgid "你好，{username}。"
msgstr "Hello, {username}."
...
```

注意：如果没有使用命令自动生成，那么很可能没有头部信息，在实际使用过程中会出现编码错误等问题，需要添加以下头部信息，完整的头部信息可以查看 [官方文档](https://www.gnu.org/software/gettext/manual/gettext.html#Header-Entry) ：

```
msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
```

### 编译文本

当我们抽取完所有文本并全部翻译后，我们就可以进行编译了，以供 Django 在运行时使用。

```
# 编译所有的 django.po 为对应的 django.mo
python3 manage.py compilemessages
```

一般本步骤直接写在 `Dockerfile` 文件内，打包时会自动处理，本地调试时会手动运行。

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

中间件实现很简单，就是获取请求头中的时区，不存在时使用全局配置中的时区；然后在当前线程中激活，待请求完毕后再清除时区信息。

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
