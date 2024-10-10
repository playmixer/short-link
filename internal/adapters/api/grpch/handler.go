package grpch

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/playmixer/short-link/internal/adapters/api/grpch/proto"
	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage/storeerror"
)

// Login - получаем токен по идентификатору.
func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (response *pb.LoginResponse, err error) {
	response = &pb.LoginResponse{}

	id := req.GetId()
	if id == "" {
		return nil, errors.Join(status.Error(codes.NotFound, "id not found"), errors.New("id not found"))
	}

	token, err := s.auth.CreateJWT(id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed create jwt, error: %v", err)
	}
	response.AccessToken = token

	return response, nil
}

// NewShort создает короткую ссылку.
func (s *Server) NewShort(ctx context.Context, req *pb.NewShortRequest) (response *pb.NewShortResponse, err error) {
	response = &pb.NewShortResponse{}

	userID, err := s.getAuth(ctx)
	if err != nil {
		return response, errors.Join(status.Error(codes.Unauthenticated, err.Error()), err)
	}

	link := strings.TrimSpace(req.GetOriginalUrl())
	_, err = url.ParseRequestURI(link)
	if err != nil {
		response.Error = fmt.Sprintf("url invalid format `%s`", link)
		return response, errors.Join(err, status.Errorf(codes.InvalidArgument, "url invalid format `%s`", link))
	}
	sLink, err := s.short.Shorty(ctx, userID, req.GetOriginalUrl())
	if err != nil {
		if errors.Is(err, storeerror.ErrNotUnique) {
			response.Short = sLink
			response.Error = fmt.Sprintf("URI `%s` already shortened", req.GetOriginalUrl())
			return response, nil
		}
		response.Error = fmt.Sprintf("failed create short url by original `%s`, error: %s", req.GetOriginalUrl(), err.Error())
		return response, errors.Join(err, status.Error(codes.Aborted, err.Error()))
	}
	response.Short = sLink
	return response, nil
}

// NewShorts - создаем список коротких ссылок.
func (s *Server) NewShorts(ctx context.Context, req *pb.NewShortsRequest) (*pb.NewShortsResponse, error) {
	response := &pb.NewShortsResponse{}

	userID, err := s.getAuth(ctx)
	if err != nil {
		return response, errors.Join(err, status.Error(codes.Unauthenticated, err.Error()))
	}

	payload := []models.ShortenBatchRequest{}
	for _, v := range req.GetOriginals() {
		_, err = url.ParseRequestURI(v.GetOriginalUrl())
		if err != nil {
			response.Error = err.Error()
			return response, errors.Join(
				err,
				status.Errorf(codes.Aborted, "url `%s` not valid, error: %s", v.GetOriginalUrl(), err.Error()),
			)
		}
		payload = append(payload, models.ShortenBatchRequest{
			CorrelationID: v.GetCorrelationId(),
			OriginalURL:   v.GetOriginalUrl(),
		})
	}

	sLink, err := s.short.ShortyBatch(ctx, userID, payload)
	for i, v := range sLink {
		sLink[i].ShortURL = v.ShortURL
		response.Shorts = append(response.Shorts, &pb.ShortenBatchResponse{
			CorrelationId: v.CorrelationID,
			ShortUrl:      v.ShortURL,
		})
	}
	if err != nil {
		if errors.Is(err, storeerror.ErrNotUnique) {
			return response, errors.Join(err, status.Error(codes.FailedPrecondition, "Conflict data"))
		}
		response.Error = err.Error()
		return response, errors.Join(err, status.Error(codes.Aborted, err.Error()))
	}

	return response, nil
}

// GetURLByShort получить оригинальную ссылку по короткой.
func (s *Server) GetURLByShort(ctx context.Context, req *pb.GetUrlByShortRequest) (*pb.GetURLByShortResponse, error) {
	response := &pb.GetURLByShortResponse{}

	link, err := s.short.GetURL(ctx, req.GetShortUrl())
	if err != nil {
		if errors.Is(err, storeerror.ErrShortURLDeleted) {
			response.Error = "URL was deleted"
			return response, errors.Join(err, status.Error(codes.NotFound, "URL was deleted"))
		}
		response.Error = err.Error()
		return response, errors.Join(err, status.Error(codes.FailedPrecondition, err.Error()))
	}

	response.OriginalUrl = link
	return response, nil
}

// GetUserURLs получить все ссылки пользователя.
func (s *Server) GetUserURLs(ctx context.Context, req *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	response := &pb.GetUserURLsResponse{}

	userID, err := s.getAuth(ctx)
	if err != nil {
		return response, errors.Join(err, status.Error(codes.Unauthenticated, err.Error()))
	}

	links, err := s.short.GetAllURL(ctx, userID)
	if err != nil {
		return response, errors.Join(err, status.Error(codes.Aborted, err.Error()))
	}

	for _, v := range links {
		response.Urls = append(response.Urls, &pb.ShortenURLs{
			ShortUrl:    v.ShortURL,
			OriginalUrl: v.OriginalURL,
		})
	}
	if len(links) == 0 {
		return response, errors.Join(errors.New("failed get urls"), status.Error(codes.NotFound, "Not found"))
	}

	return response, nil
}

// DeleteUserURLs удалить ссылки пользователя.
func (s *Server) DeleteUserURLs(ctx context.Context, req *pb.DeleteUserURLsRequest) (*pb.DeleteUserURLsRespons, error) {
	response := &pb.DeleteUserURLsRespons{}

	userID, err := s.getAuth(ctx)
	if err != nil {
		return response, errors.Join(err, status.Error(codes.Unauthenticated, err.Error()))
	}

	data := []models.ShortLink{}
	for _, short := range req.GetShortUrls() {
		data = append(data, models.ShortLink{UserID: userID, ShortURL: short})
	}

	err = s.short.DeleteShortURLs(ctx, data)
	if err != nil {
		return response, errors.Join(err, status.Error(codes.Aborted, err.Error()))
	}
	return response, nil
}

// GetStatus статистика сохраненных ссылок.
func (s *Server) GetStatus(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
	response := &pb.GetStatusResponse{}

	access := true
	network, err := netip.ParsePrefix(s.trustedSubnet)
	if err != nil {
		s.log.Debug("trusted subnet is not valid", zap.Error(err), zap.String("subnet", s.trustedSubnet))
		access = false
	}
	ipStr, err := s.getMetadata(ctx, "X-Real-IP")
	if err != nil {
		return response, status.Errorf(http.StatusForbidden, "not valid real ip `%s`", ipStr)
	}
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		s.log.Debug("IP address is not valid", zap.Error(err), zap.String("ip", ipStr))
		access = false
	}
	if ok := network.Contains(ip); !ok {
		access = false
	}

	if !access {
		return response, errors.Join(errors.New("forbidden"), status.Error(http.StatusForbidden, "forbidden"))
	}

	stats, err := s.short.GetState(ctx)
	if err != nil {
		return response, errors.Join(err, status.Errorf(codes.Aborted, "failed to get state, error: %s", err.Error()))
	}
	response.Urls = int32(stats.URLs)
	response.Users = int32(stats.Users)
	return response, nil
}
