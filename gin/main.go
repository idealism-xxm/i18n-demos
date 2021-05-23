package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.elastic.co/apm/module/apmgin"
	"go.elastic.co/apm/module/apmgrpc"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	gingrpc "gin-i18n/grpc"
	"gin-i18n/i18n"
)

func main() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	go func() {
		startGinServer()
		waitGroup.Done()
	}()
	time.Sleep(time.Second)
	go func() {
		startGrpcServer()
		waitGroup.Done()
	}()
	waitGroup.Wait()
}

func startGinServer() {
	r := gin.Default()
	r.Use(apmgin.Middleware(r))
	r.Use(i18n.GinMiddleware())
	r.Use(TzMiddleware())

	r.GET("/hello/:username/", func(c *gin.Context) {
		// 获取路径中的 username
		username := c.Param("username")
		// 从 context 中获取 languageTag 和 location
		ctx := c.Request.Context()
		langaugeTag := i18n.LanguageTagFromContext(ctx)
		location := LocationFromContext(ctx)
		currentLanguage := i18n.CurrentLanguage(ctx, langaugeTag.String())
		personCat := i18n.PersonCats(ctx, username, 1)
		personCats := i18n.PersonCats(ctx, username, 2)
		nowStr := time.Now().In(location).Format(time.RFC3339Nano)
		timeStr := fmt.Sprintf("(%s) %s", location.String(), nowStr)
		c.String(http.StatusOK, fmt.Sprintf("%v\n%v\n%v\n%v", currentLanguage, personCat, personCats, timeStr))
	})

	if err := r.Run(); err != nil {
		panic(err)
	}
}

type grpcServer struct {
	gingrpc.UnimplementedGinServiceServer
}

func (*grpcServer) Hello(ctx context.Context, req *gingrpc.HelloRequest) (*gingrpc.HelloResponse, error) {
	return &gingrpc.HelloResponse{
		Message: fmt.Sprintf("Hello %s.", req.Name),
	}, nil
}

func startGrpcServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_ctxtags.UnaryServerInterceptor(),
			apmgrpc.NewUnaryServerInterceptor(),
		),
	)
	gingrpc.RegisterGinServiceServer(s, &grpcServer{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
