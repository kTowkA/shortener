package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	pb "github.com/kTowkA/shortener/internal/grpc/proto"
	"github.com/kTowkA/shortener/internal/storage"
	"github.com/kTowkA/shortener/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ShortenerServer наше приложение для реализации gRPC сервиса Shortener
type ShortenerServer struct {
	pb.UnimplementedShortenerServer
	db     storage.Storager
	logger *slog.Logger
}

// CreategRPCServer создает структуру реализующую gRPC сервис Shortener которую будем регистрировать
func CreategRPCServer(db storage.Storager, logger *slog.Logger) *ShortenerServer {
	return &ShortenerServer{
		db:     db,
		logger: logger,
	}
}

// DecodeURL реализация gRPC сервиса Shortener
func (s *ShortenerServer) DecodeURL(ctx context.Context, r *pb.DecodeURLRequest) (*pb.DecodeURLResponse, error) {
	resp, err := s.db.RealURL(ctx, r.ShortUrl)
	if errors.Is(err, storage.ErrURLNotFound) {
		s.logger.Debug("поиск оригинального URL. ничего не найдено", slog.String("short", r.ShortUrl))
		return nil, status.Error(codes.NotFound, fmt.Sprintf("по запросу \"%s\" ничего не найдено", r.ShortUrl))
	} else if err != nil {
		s.logger.Error("поиск оригинального URL", slog.String("short", r.ShortUrl), slog.String("ошибка", err.Error()))
		return nil, err
	}
	if resp.IsDeleted {
		s.logger.Debug("поиск оригинального URL. ресурс удален", slog.String("short", r.ShortUrl))
		return nil, fmt.Errorf("ресурс \"%s\" уже был удален", r.ShortUrl)
	}
	return &pb.DecodeURLResponse{OriginalUrl: resp.OriginalURL}, nil
}

// EncodeURL реализация gRPC сервиса Shortener
func (s *ShortenerServer) EncodeURL(ctx context.Context, r *pb.EncodeURLRequest) (*pb.EncodeURLResponse, error) {
	if _, err := url.Parse(r.OriginalUrl); err != nil {
		s.logger.Error("сокращение URL", slog.String("short", r.OriginalUrl), slog.String("ошибка", err.Error()))
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("переданное значение \"%s\" не явялется валидной ссылкой", r.OriginalUrl))
	}
	// здесь можно было обойтись без выхода в случае отсутствия userID, но пусть будет так. С новой сокращенной ссылкой всегда должен быть создавший ее пользователь
	userID, err := userIDFromContext(ctx)
	if err != nil {
		s.logger.Error("получение ID пользователя", slog.String("ошибка", err.Error()))
		return nil, err
	}

	short, err := utils.SaveLink(ctx, s.db, userID, r.OriginalUrl)
	if errors.Is(err, storage.ErrURLConflict) {
		s.logger.Debug("сокращение URL. конфликт", slog.String("short", r.OriginalUrl))
		return &pb.EncodeURLResponse{SavedLink: short, Error: storage.ErrURLConflict.Error()}, nil
	}
	if err != nil {
		s.logger.Error("сокращение URL", slog.String("short", r.OriginalUrl), slog.String("ошибка", err.Error()))
		return nil, err
	}
	return &pb.EncodeURLResponse{
		SavedLink: short,
	}, nil
}

// Batch реализация gRPC сервиса Shortener
func (s *ShortenerServer) Batch(ctx context.Context, r *pb.BatchRequest) (*pb.BatchResponse, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		s.logger.Error("получение ID пользователя", slog.String("ошибка", err.Error()))
		return nil, err
	}
	batch := batchRequestToModelBatchRequest(r)
	resp, err := utils.SaveBatch(ctx, s.db, userID, batch)
	if err != nil {
		s.logger.Error("сохранение массива значений", slog.String("ошибка", err.Error()))
		return nil, err
	}
	return modelBatchResponseToBatchResponse(resp), nil
}

// UserURLs реализация gRPC сервиса Shortener
func (s *ShortenerServer) UserURLs(ctx context.Context, r *pb.UserURLsRequest) (*pb.UserURLsResponse, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		s.logger.Error("получение ID пользователя", slog.String("ошибка", err.Error()))
		return nil, err
	}
	resp, err := s.db.UserURLs(ctx, userID)
	if errors.Is(err, storage.ErrURLNotFound) {
		s.logger.Debug("получение ссылок пользователя. ничего не найдено")
		return nil, status.Error(codes.NotFound, storage.ErrURLNotFound.Error())
	}
	if err != nil {
		s.logger.Error("получение ссылок пользователя", slog.String("ошибка", err.Error()))
		return nil, err
	}
	return modelStorageJSONToUserURLsResponse(resp), nil
}

// DeleteUserURLs реализация gRPC сервиса Shortener
func (s *ShortenerServer) DeleteUserURLs(ctx context.Context, r *pb.DelUserRequest) (*pb.DeleteUserURLsResponse, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		s.logger.Error("получение ID пользователя", slog.String("ошибка", err.Error()))
		return nil, err
	}
	err = s.db.DeleteURLs(ctx, delUserRequestToModelDeleteURLMessage(userID.String(), r))
	if err != nil {
		s.logger.Error("удаление ссылок пользователя", slog.String("ошибка", err.Error()))
		return nil, err
	}
	return &pb.DeleteUserURLsResponse{}, nil
}

// Stats реализация gRPC сервиса Shortener
func (s *ShortenerServer) Stats(ctx context.Context, r *pb.StatsRequest) (*pb.StatsResponse, error) {
	stats, err := s.db.Stats(ctx)
	if err != nil {
		s.logger.Error("получение статистики", slog.String("ошибка", err.Error()))
		return nil, err
	}
	return &pb.StatsResponse{Users: int32(stats.TotalUsers), Urls: int32(stats.TotalURLs)}, nil
}

// Ping реализация gRPC сервиса Shortener
func (s *ShortenerServer) Ping(ctx context.Context, r *pb.PingRequest) (*pb.PingResponse, error) {
	err := s.db.Ping(ctx)
	if err != nil {
		s.logger.Error("проверка доступности сервиса", slog.String("ошибка", err.Error()))
		return nil, status.Error(codes.Unavailable, "в настоящее время сервис не доступен")
	}
	return &pb.PingResponse{Status: &pb.PingResponse_Status{Ok: true}}, nil
}
