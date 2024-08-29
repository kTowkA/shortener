package server

import (
	pb "github.com/kTowkA/shortener/internal/grpc/proto"
	"github.com/kTowkA/shortener/internal/model"
	"github.com/kTowkA/shortener/internal/utils"
)

func batchRequestToModelBatchRequest(r *pb.BatchRequest) model.BatchRequest {
	batch := make([]model.BatchRequestElement, 0, len(r.Elements))
	for _, value := range r.Elements {
		batch = append(batch, model.BatchRequestElement{
			CorrelationID: value.CorrelationId,
			OriginalURL:   value.OriginalUrl,
		})
	}
	return utils.ValidateAndGenerateBatch(batch)
}
func modelBatchResponseToBatchResponse(r model.BatchResponse) *pb.BatchResponse {
	batch := make([]*pb.BatchResponse_Result, 0, len(r))
	for i := range r {
		batch = append(batch, &pb.BatchResponse_Result{
			CorrelationId: r[i].CorrelationID,
			ShortUrl:      r[i].ShortURL,
		})
	}
	return &pb.BatchResponse{Result: batch}
}
func modelStorageJSONToUserURLsResponse(r []model.StorageJSON) *pb.UserURLsResponse {
	result := make([]*pb.UserURLsResponse_Result, 0, len(r))
	for i := range r {
		result = append(result, &pb.UserURLsResponse_Result{
			ShortUrl:    r[i].ShortURL,
			OriginalUrl: r[i].OriginalURL,
			IsDeleted:   r[i].IsDeleted,
			Uuid:        r[i].UUID,
		})
	}
	return &pb.UserURLsResponse{Result: result}
}

func delUserRequestToModelDeleteURLMessage(userID string, r *pb.DelUserRequest) []model.DeleteURLMessage {
	result := make([]model.DeleteURLMessage, 0, len(r.ShortUrls))
	for i := range r.ShortUrls {
		result = append(result, model.DeleteURLMessage{
			UserID:   userID,
			ShortURL: r.ShortUrls[i],
		})
	}
	return result
}
