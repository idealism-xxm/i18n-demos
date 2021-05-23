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
# 配置 APM 所需的环境变量
ENV ELASTIC_APM_SERVICE_NAME gin-i18n
ENV ELASTIC_APM_SERVER_URL http://host.docker.internal:8200
ENV ELASTIC_APM_ENVIRONMENT local

ENTRYPOINT ["/idealism-xxm/gin/gin-i18n"]
CMD ["run", "--logtostderr"]
