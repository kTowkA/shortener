package app

import (
	"context"
	"log/slog"
	"net"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	pb "github.com/kTowkA/shortener/internal/grpc/proto"
	"github.com/kTowkA/shortener/internal/grpc/server"
	"github.com/kTowkA/shortener/internal/storage"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Run запуск gRPC сервера
func Run(ctx context.Context, db storage.Storager, log *slog.Logger, address string) error {

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(),
		userID,
	))

	gr, grCtx := errgroup.WithContext(ctx)

	gr.Go(func() error {
		defer log.Info("остановили сервер")
		<-grCtx.Done()
		gRPCServer.GracefulStop()
		return nil
	})
	gr.Go(func() error {
		s := server.CreategRPCServer(db, log)
		pb.RegisterShortenerServer(gRPCServer, s)

		l, err := net.Listen("tcp", address)
		if err != nil {
			return err
		}

		log.Info("gRPC server started")

		if err := gRPCServer.Serve(l); err != nil {
			return err
		}
		return nil
	})
	return gr.Wait()
}

// userID такой искуственный пример перехватчика. если нет id пользователя, то генерируем новый и сохраняем в контексте
// можно было вынести в общий код работу с jwt токеном, но он что там был бесполезен, так как он созхдавался и в resp api автоматом, поэтому быстрый вариант показать что умею и перехватчики
func userID(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("userid")
		if _, err := uuid.Parse(values[0]); err == nil {
			return handler(ctx, req)
		}
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("userid", uuid.New().String()))
	return handler(ctx, req)
}
