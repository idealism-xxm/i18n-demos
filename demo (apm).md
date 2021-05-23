## 简介

在 [Django i18n](django/demo%20(simple%20i18n).md) 和 [Gin i18n](gin/demo%20(simple%20i18n).md) 中分别介绍了如何在 Django 和 Gin 中实现国际化和本地化，但未涉及如何实现全链路国际化和本地化。

要做到全链路，就需要将对应的信息进行跨服务传递，最常见的场景就是全链路追踪。本次将介绍如何通过 gRPC 在 Python 和 Golang 的服务间传递 traceId ，从而实现全链路追踪。

## 搭建本地 eak 环境

我们将使用 Elasticsearch + Apm + Kibana 进行演示，所以先需要搭建本地 eak 环境。

对应的三个组件都支持 Docker 部署，所以我们可以直接使用一个如下的 `docker-compose-eak.yaml` 直接搭建完成。

```yaml
# 搭建 elasticsearch + apm + kibana 服务
version: '3.6'
services:
  elasticsearch:
    image: elasticsearch:7.12.1
    container_name: elasticsearch
    ports:
      - 9200:9200
    networks:
      - eak
    environment:
      - discovery.type=single-node
    healthcheck:
      test: curl --cacert /usr/share/elasticsearch/config/certs/ca/ca.crt -s https://localhost:9200 >/dev/null; if [[ $$? == 52 ]]; then echo 0; else echo 1; fi
      interval: 30s
      timeout: 10s
      retries: 5

  kibana:
    image: kibana:7.12.1
    container_name: kibana
    ports:
      - 5601:5601
    networks:
      - eak
    depends_on:
      - elasticsearch
    healthcheck:
      test: curl --cacert /usr/share/elasticsearch/config/certs/ca/ca.crt -s https://localhost:5601 >/dev/null; if [[ $$? == 52 ]]; then echo 0; else echo 1; fi
      interval: 30s
      timeout: 10s
      retries: 5

  apm-server:
    image: docker.elastic.co/apm/apm-server:7.12.1
    container_name: apm_server
    ports:
      - 8200:8200
    networks:
      - eak
    # -e flag to log to stderr and disable syslog/file output
    command: --strict.perms=false -e
    depends_on:
      - elasticsearch
      - kibana
    healthcheck:
      test: curl --cacert /usr/share/elasticsearch/config/certs/ca/ca.crt -s https://localhost:8200/healthcheck >/dev/null; if [[ $$? == 52 ]]; then echo 0; else echo 1; fi
      interval: 30s
      timeout: 10s
      retries: 5

networks:
  eak: {}
# use docker volume to persist ES data outside of a container.
volumes:
  es_data: {}
```

运行如下命令启动 eak 环境：

```shell
docker-compose -f docker-compose-eak.yml up
```

## Django 中接入 APM

先安装 elastic-apm 依赖，本示例使用 poetry 管理依赖，使用如下命令添加然后重新打包镜像即可：

```shell
# 安装依赖
poetry add elastic-apm
# 重新打包镜像
make build-django
```

此时我们可以在代码中进行配置了，修改 `settings.py` 文件中的如下内容（没有则添加）：：

```python
# 添加代理对应的 app 到 installed apps 中
INSTALLED_APPS = (
  ...
  'elasticapm.contrib.django',
)

# 配置 APM 相关信息
ELASTIC_APM = {
  # 配置服务名，支持如下字符：
  # a-z, A-Z, 0-9, -, _, and space
  'SERVICE_NAME': 'django-i18n',

  # 配置 token ，本地环境忽略
  'SECRET_TOKEN': '',

  # 配置 APM 服务器的链接，在 docker 中，使用 host.docker.internal 替代 localhost
  # 默认链接： http://localhost:8200
  'SERVER_URL': 'http://host.docker.internal:8200',

  # 配置服务的环境，本地环境设为 local
  'ENVIRONMENT': 'local',
  
  # 设置为 True ，可以在 Django 的 Debug 模式下也采集数据
  # 默认配置： False
  'DEBUG': True,
}

# 加上追踪中间件
MIDDLEWARE = (
  ...
  'elasticapm.contrib.django.middleware.TracingMiddleware',
  ...
)
```

以上是必要的简单配置，如果有更多的定制化需求，可以查看 [官方文档](https://www.elastic.co/guide/en/apm/agent/python/current/configuration.html) 进行详细配置。

配置好后，我们就可以先访问 [Django 服务](http://localhost:8000/hello/idealism/) 以便 APM 采集到数据。

然后访问 [Kibana 图形化界面](http://localhost:5601/app/apm/services) 即可看到我们配置好的服务已经采集到了：

![Kibana Apm Services Django](img/Kibana%20Apm%20Services%20Django.png)

点击 django-i18n 即可查看 Django 服务采集的所有数据，我们刚刚访问的请求被全部采集到了：

![Kibana Apm django-i18n](img/Kibana%20Apm%20django-i18n.png)

剩下的都是常见的一些操作，这里就先不展示了。 elastic-apm 的 Django 客户端已经能自动采集大部分我们会使用到的服务：mysql, redis, http 等。如果需要额外的采集信息，可以自己实现相关逻辑即可。

## Gin 中接入 APM

Golang 中使用 APM 就不是那么简单了，虽然官方库提供了很多组件的客户端，但是有些组件的不同版本不兼容，需要自己重新实现对应的采集逻辑。

常见组件的客户端基本都有，所以即使遇到版本不兼容的问题，也可以很轻易自定义使用版本的客户端，我们实际使用时就自定义了好几个组件对应版本的客户端。实现方式大同小异，要么直接使用组件自带的方式（中间件、拦截器、Hook 等）进行处理，要么封装一层进行处理。

Golang 中的 APM 通过环境变量的方式进行配置，增加以下必要的配置即可正常使用：

```shell
# 配置服务名，支持如下字符：
# a-z, A-Z, 0-9, -, _, and space
# 默认值：可执行文件名称
ELASTIC_APM_SERVICE_NAME=gin-i18n

# 配置 token ，本地环境忽略
ELASTIC_APM_SECRET_TOKEN=

# 配置 APM 服务器的链接，在 docker 中，使用 host.docker.internal 替代 localhost
# 默认链接： http://localhost:8200
ELASTIC_APM_SERVER_URL=http://host.docker.internal:8200

# 配置服务的环境，本地环境设为 local
ELASTIC_APM_ENVIRONMENT=local
```

如有更多定制化的需求，可以查看 [官方文档](https://www.elastic.co/guide/en/apm/agent/go/current/configuration.html) 进行详细配置。

这里先只接入 Gin 客户端，中间件的方式非常简单，其他的接入方式也与组件自身提供的功能有关。

```go
func main() {
	...
    r := gin.Default()
    apmgin.Middleware(r)
    ...
}
```

这样配置完成后，我们就可以先访问 [Gin 服务](http://localhost:8080/hello/idealism/) 以便 APM 采集到数据。

然后访问 [Apm gin-i18n](http://localhost:5601/app/apm/services/gin-i18n/overview) 即可查看收集到的数据：

![Kibana APM gin-i18n](img/Kibana%20APM%20gin-i18n.png)

## 完成 gRPC 的全链路追踪

到目前为止，我们分别在 Django 和 Gin 中实现了分离的追踪，但无法在微服务场景下串起来客户端和服务端的关系，难以快速在微服务场景下进行排查。

此时就需要全链路追踪技术，并且利用全链路追踪的 traceId ，我们还可以一次将一条链路上的所有日志全部拉出来。我们可以同时结合链路上的追踪记录和日志，这样极大地提高排查问题的效率，找到问题所在。

### 分析 Golang gRPC 的 Apm 客户端

Apm 官方库已经提供了 Golang 的 gRPC 的客户端，而尚未提供 Python 的 gRPC 的客户端，所以我们先分析 Golang gRPC 的 Apm 客户端，看看是如何传递 traceId 的。

```go
// go.elastic.co/apm/module/apmgrpc@v1.11.0/server.go:97
func startTransaction(ctx context.Context, tracer *apm.Tracer, name string) (*apm.Transaction, context.Context) {
    var opts apm.TransactionOptions
    if md, ok := metadata.FromIncomingContext(ctx); ok {
    	// elasticTraceparentHeader 的值为 Elastic-Apm-Traceparent
        traceContext, ok := getIncomingMetadataTraceContext(md, elasticTraceparentHeader)
        if !ok {
        	// w3cTraceparentHeader 的值为 Traceparent
            traceContext, _ = getIncomingMetadataTraceContext(md, w3cTraceparentHeader)
        }
        opts.TraceContext = traceContext
    }
    tx := tracer.StartTransactionOptions(name, "request", opts)
    tx.Context.SetFramework("grpc", grpc.Version)
    return tx, apm.ContextWithTransaction(ctx, tx)
}
```

apmgrpc 会从 grpc 的 metadata 中获取获取 trace 的相关信息，优先获取 `Elastic-Apm-Traceparent` 指定的 traceId ，没有获取到会再获取 `Traceparent` 指定的 traceId 。由于后者是 W3C 标准的请求头，所以我们决定采用 `Traceparent` 传递 traceId 。

如果需要自定义其他的请求头名，可以使用 grpc_opentracing 提供的拦截器。

### gRPC-go 服务端支持 apm

想要让 gRPC 的 Golang 服务端支持 apm ，我们可以直接在创建新服务端实例时添加上 apmgrpc 的拦截器即可：

```go
s := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        grpc_recovery.UnaryServerInterceptor(),
        grpc_ctxtags.UnaryServerInterceptor(),
        apmgrpc.NewUnaryServerInterceptor(),
    ),
)
```

### 编写 gRPC-python 客户端的 Apm 拦截器

现在我们已经知道了需要通过 `Traceparent` 请求头传递 traceId ，那么我们可以继续编写 Python 下的 gRPC 客户端的拦截器了，在每次请求的时候都将当前的 traceId 放入 metadata 即可。

## 小结

