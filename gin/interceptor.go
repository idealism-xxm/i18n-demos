package main

import (
	"context"
	"gin-i18n/i18n"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

const (
	grpcHeaderNameLanguage = "x-language"
	grpcHeaderNameTimezone = "x-timezone"
)

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
