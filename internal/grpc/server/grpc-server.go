package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/google/uuid"
	"github.com/kTowkA/shortener/internal/storage"
	"github.com/kTowkA/shortener/internal/utils"
	pb "github.com/kTowkA/shortener/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ShortenerServer struct {
	pb.UnimplementedShortenerServer
	db storage.Storager
}

func (s *ShortenerServer) DecodeURL(ctx context.Context, r *pb.DecodeURLRequest) (*pb.DecodeURLResponse, error) {
	resp, err := s.db.RealURL(ctx, r.ShortUrl)
	if errors.Is(err, storage.ErrURLNotFound) {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("по запросу \"%s\" ничего не найдено", r.ShortUrl))
	} else if err != nil {
		return nil, err
	}
	if resp.IsDeleted {
		return nil, fmt.Errorf("ресурс \"%s\" уже был удален", r.ShortUrl)
	}
	return &pb.DecodeURLResponse{OriginalUrl: resp.OriginalURL}, nil
}
func (s *ShortenerServer) EncodeURL(ctx context.Context, r *pb.EncodeURLRequest) (*pb.EncodeURLResponse, error) {
	if _, err := url.Parse(r.OriginalUrl); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("переданное значение \"%s\" не явялется валидной ссылкой", r.OriginalUrl))
	}
	userID, err := uuid.Parse(r.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("переданное значение \"%s\" не явялется валидным uuid", r.UserId))
	}
	short, err := utils.SaveLink(ctx, s.db, userID, r.OriginalUrl)
	log.Println(short, err)
	if errors.Is(err, storage.ErrURLConflict) {
		return &pb.EncodeURLResponse{SavedLink: short, Error: storage.ErrURLConflict.Error()}, nil
	}
	if err != nil {
		return nil, err
	}
	return &pb.EncodeURLResponse{
		SavedLink: short,
	}, nil
}
func (s *ShortenerServer) Batch(context.Context, *pb.BatchRequest) (*pb.BatchResponse, error) {
	return &pb.BatchResponse{}, nil
}
func (s *ShortenerServer) UserURLs(context.Context, *pb.UserURLsRequest) (*pb.UserURLsResponse, error) {
	return &pb.UserURLsResponse{}, nil
}
func (s *ShortenerServer) DeleteUserURLs(context.Context, *pb.DelUserRequest) (*pb.DeleteUserURLsResponse, error) {
	return &pb.DeleteUserURLsResponse{}, nil
}
func (s *ShortenerServer) Stats(ctx context.Context, r *pb.StatsRequest) (*pb.StatsResponse, error) {
	stats, err := s.db.Stats(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.StatsResponse{Users: int32(stats.TotalUsers), Urls: int32(stats.TotalURLs)}, nil
}
func (s *ShortenerServer) Ping(ctx context.Context, r *pb.PingRequest) (*pb.PingResponse, error) {
	err := s.db.Ping(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, "в настоящее время сервис не доступен")
	}
	return &pb.PingResponse{Status: &pb.PingResponse_Status{Ok: true}}, nil
}
