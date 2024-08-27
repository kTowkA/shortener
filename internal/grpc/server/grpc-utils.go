package server

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	keyUserID = "userid"
)

func userIDFromContext(ctx context.Context) (uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.UUID{}, status.Error(codes.Unauthenticated, "ID пользователя не содержится в запросе")
	}
	values := md.Get(keyUserID)
	if len(values) == 0 {
		return uuid.UUID{}, status.Error(codes.Unauthenticated, "ID пользователя не содержится в запросе")
	}
	userID, err := uuid.Parse(values[0])
	if err != nil {
		return uuid.UUID{}, status.Error(codes.InvalidArgument, fmt.Sprintf("переданное значение \"%s\" не явялется валидным uuid", values[0]))
	}
	return userID, nil
}
