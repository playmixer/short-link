package grpch

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (s *Server) interceptorLogger(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (
	interface{}, error,
) {
	start := time.Now()
	defer func() {
		s.log.Info("[INFO]", zap.String("method", info.FullMethod), zap.Duration("duration", time.Since(start)))
	}()
	return handler(ctx, req)
}
