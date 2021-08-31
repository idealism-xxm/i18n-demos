## 简介

在 [Django i18n](django/demo%20(simple%20i18n).md) 和 [Gin i18n](gin/demo%20(simple%20i18n).md) 中分别介绍了如何在 Django 和 Gin 中实现国际化和本地化，并在 [gRPC APM](demo%20(apm).md) 中介绍了如何实现 gRPC 的全链路追踪。

现在我们就要组合前面介绍的知识，实现我们所需要的全链路国际化和本地化功能（当然可以扩展至任意需要全链路传递的信息）。

本次将实现 gRPC 上的全链路国际化，全链路追踪和 Elasticsearch + Apm + Kibana 环境搭建可以直接查看 [gRPC APM](demo%20(apm).md) 。

## gRPC-python 客户端适配

### 抽象 gRPC-python 的拦截器

我们已经实现过 gRPC-python 的 APM 拦截器，并且其中大部分都是可以复用的模版代码。

不同拦截器只有获取需要传递的 metadata 列表不同，那么我们可以实现一个模版方法模式，将获取 metadata 列表的逻辑抽取出来，让子类只关心这部分即可。

```python
class CustomUnaryUnaryClientInterceptor(UnaryUnaryClientInterceptor):
    """
    自定义一元客户端拦截器
    使用模版方法模式，简化子类加入 metadata 信息
    """
    def intercept_unary_unary(self, continuation, client_call_details, request):
        # 先获取原有的 metadata
        metadata = []
        if client_call_details.metadata is not None:
            metadata = list(client_call_details.metadata)

        # 获取子类中得到的 metadata 信息列表，放入到 metadata 中
        metadata.extend(
            self._get_metadata_list(continuation, client_call_details, request)
        )

        # 生成最新的 details
        new_details = _ClientCallDetails(
            client_call_details.method,
            client_call_details.timeout,
            metadata,
            client_call_details.credentials,
            client_call_details.wait_for_ready,
            client_call_details.compression,
        )

        return continuation(new_details, request)

    def _get_metadata_list(self, continuation, client_call_details, request) -> List[Tuple[str, str]]:
        """获取 metadata 信息列表（子类重写）"""
        return []
```

### 实现 gRPC-python 的语言和时区拦截器

语言信息和时区信息我们已经在 [Django i18n](django/demo%20(simple%20i18n).md) 中讲解过如何从请求中获取，我们这里直接获取并返回即可，如果没有会自动获取我们预定义的默认值。

```python
# 定义服务端从这些头中获取相应信息
GRPC_HEADER_NAME_TIMEZONE = 'x-timezone'
GRPC_HEADER_NAME_LANGUAGE = 'x-language'


class LanguageUnaryUnaryClientInterceptor(CustomUnaryUnaryClientInterceptor):
    def _get_metadata_list(self, continuation, client_call_details, request) -> List[Tuple[str, str]]:
        # 返回 语言 的相关信息
        return [(GRPC_HEADER_NAME_LANGUAGE, trans_real.get_language())]


class TimezoneUnaryUnaryClientInterceptor(CustomUnaryUnaryClientInterceptor):
    def _get_metadata_list(self, continuation, client_call_details, request) -> List[Tuple[str, str]]:
        # 返回 时区 的相关信息
        return [(GRPC_HEADER_NAME_TIMEZONE, timezone.get_current_timezone_name())]
```

可以发现样板代码已被消除，子类中只需要实现必须的关键代码。

### 运用 gRPC-python 的拦截器

接下来就是在使用时，拿到 channel 后调用 interceptor_channel 函数将所有的拦截器串起来即可。

由于追踪、语言和时区这三个信息每次请求必定传递，所以我们可以实现一个简单的封装，方便调用方直接使用。

```
trace_unary_unary_client_interceptor = TraceUnaryUnaryClientInterceptor()
language_unary_unary_client_interceptor = LanguageUnaryUnaryClientInterceptor()
timezone_unary_unary_client_interceptor = TimezoneUnaryUnaryClientInterceptor()


def intercept_channel_all(channel):
    """应用所有的拦截器"""
    return grpc.intercept_channel(
        channel,
        trace_unary_unary_client_interceptor,
        language_unary_unary_client_interceptor,
        timezone_unary_unary_client_interceptor,
    )
```

然后我们在需要使用的地方直接调用，将这三个信息全部附加到请求中，服务端即可接收到这些信息。

```python
def hello(name: str) -> str:
    with grpc.insecure_channel('host.docker.internal:50051') as channel:
        # 对 channel 应用所有拦截器
        channel = interceptor.intercept_channel_all(channel)
        stub = GinServiceStub(channel)
        # 如果不想使用拦截器，也可以在调用的时候指定所需信息的 metadata
        response = stub.Hello(HelloRequest(name=name))

    return response.message
```

至此我们已经完成了客户端方面的改造，无需改动业务代码即可完成所有信息的传递。

## gRPC-go 服务端适配

### 实现 gRPC-go 的时区拦截器

现在我们需要实现服务端的拦截器，将时区信息解析出来，并放入到 ctx 中，使得业务代码可以直接使用。

这部分也很简单，先从 ctx 中获取 gRPC 的 metadata ，然后再从其中获取时区信息，最后转换成对应的时区并放入到 ctx 中即可。

```go
func TimezoneUnaryServerInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (resp interface{}, err error) {
    // 从 gRPC 头中获取时区标识
    timezoneName := ""
    if md, ok := metadata.FromIncomingContext(ctx); ok {
        if values := md.Get(grpcHeaderNameTimezone); len(values) == 1 {
            timezoneName = values[0]
        }
    }
    
    // 获取时区
    location, err := time.LoadLocation(timezoneName)
    if timezoneName == "" || err != nil {
        // 没有指定时区 或者 出错，则使用默认时区
        location = defaultLocation
    }
    
    // 放入 context 中
    ctx = WithLocation(ctx, location)
    
    return handler(ctx, req)
}
```

### 实现 gRPC-go 的语言拦截器

语言信息的解析稍微复杂一点，不过原理还是和从 Gin 请求中解析类似，我们可以先将 Gin 中间件中的这部分逻辑抽取出来。

gRPC 传递过来的语言信息是 Gin 请求头中的语言信息的子集，所以可以直接复用原有逻辑。

我们这里定义一个新的函数，先解析出指定语言信息最适合的一个语言及其 tag ，然后再将语言和 tag 放入到 ctx 中并返回。

```go
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
```

然后我们可以直接用这个抽取出的函数替换原有 Gin 中的逻辑，也可以直接用于 gRPC 的语言拦截器中。

```go
func LanguageUnaryServerInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (resp interface{}, err error) {
    // 从 gRPC 头中获取语言标识
    languageCode := ""
    if md, ok := metadata.FromIncomingContext(ctx); ok {
        if values := md.Get(grpcHeaderNameLanguage); len(values) == 1 {
            languageCode = values[0]
        }
    }

    // 将语言信息放入到 ctx 中
    ctx = i18n.WithLanguageAndTag(ctx, languageCode)

    return handler(ctx, req)
}
```

### 修改 gRPC-go 的接口

最后将 hello 的 gRPC 接口进行改造，使其能够读取 ctx 中的语言和时区，并能根据这些信息返回相应的字符串，以供测试。

```go
func (*grpcServer) Hello(ctx context.Context, req *gingrpc.HelloRequest) (*gingrpc.HelloResponse, error) {
    // 从 context 中获取 languageTag 和 location
    langaugeTag := i18n.LanguageTagFromContext(ctx)
    location := LocationFromContext(ctx)
    currentLanguage := i18n.CurrentLanguage(ctx, langaugeTag.String())
    personCat := i18n.PersonCats(ctx, req.Name, 1)
    personCats := i18n.PersonCats(ctx, req.Name, 2)
    nowStr := time.Now().In(location).Format(time.RFC3339Nano)
    timeStr := fmt.Sprintf("(%s) %s", location.String(), nowStr)
    message := fmt.Sprintf("%v<br/>%v<br/>%v<br/>%v", currentLanguage, personCat, personCats, timeStr)
    return &gingrpc.HelloResponse{
        Message: message,
    }, nil
}
```

## 测试

我们可以访问 `http://localhost:8000/hello-with-grpc/idealism/` 进行测试。

不同请求头测试结果如下（可以使用 [ModHeader](https://chrome.google.com/webstore/detail/modheader/idgpnmonknjnojddfkpgkljpfnnfcklj) 插件在浏览器中快速修改请求头）：

```
# Accept-Language: zh-CN,zh;q=0.9,en;q=0.8
# x-timezone: 
当前语言：zh-Hans
idealism 有 1 只猫。
idealism 有 2 只猫。
(UTC) 2021-08-31T14:50:39.0770857Z

# Accept-Language: en
# x-timezone: Europe/Berlin
Current Language: en-US
idealism have 1 cat.
idealism have 2 cats.
(Europe/Berlin) 2021-08-31T16:53:02.4734278+02:00

# Accept-Language: en
# x-timezone: Asia/Shanghai
Current Language: en-US
idealism have 1 cat.
idealism have 2 cats.
(Asia/Shanghai) 2021-08-31T22:53:52.660427+08:00
```

这个简单的测试说明 Django 中能从请求头中正确获取到语言和时区信息，并能正确通过我们刚刚实现的拦截器传递给服务端。

## 小结

本文主要讲解了如何在 Python 和 Golang 中实现 gRPC 上的全链路国际化和本地化。

实现方式与全链路追踪一致，通过 gRPC 的 metadata 传递我们需要的信息即可。

我们简单修改 [gRPC APM](demo%20(apm).md) 中的拦截器代码，抽取公共逻辑后，就实现了前文中提到的公共组件，简化了传递语言和时区信息的逻辑。

通过 gRPC 的拦截器，我们不侵入业务逻辑即可实现全链路国际化，极大地降低了接入改造的成本。

下次将介绍 Python 的 Celery 中使用装饰器实现的全链路国际化（语言翻译和本地时区），得益于 Python 强大的扩展性，我们依旧可以不侵入业务代码。
 
