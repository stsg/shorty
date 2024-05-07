package app

import (
	"context"
	"errors"
	"fmt"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"

	pb "github.com/stsg/shorty/api"
	"github.com/stsg/shorty/internal/logger"
	"github.com/stsg/shorty/internal/storage"
)

// GRPCServer is a struct that holds gRPC server data.
type GRPCServer struct {
	pb.UnimplementedShortenerServer
	grpcServer *grpc.Server
}

// NewGRPCServer creates a new instance of the GRPCServer struct.
//
// It initializes the GRPCServer with the provided interceptors and registers the
// ShortenerServer with the gRPC server.
//
// Returns a pointer to the GRPCServer instance.
func NewGRPCServer() *GRPCServer {
	interceptors := []grpc.UnaryServerInterceptor{
		GRPCRequestLogger,
	}

	// var srv *grpc.Server
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(interceptors...)),
	)
	pb.RegisterShortenerServer(srv, &GRPCServer{
		grpcServer: srv,
	})

	// return srv

	return &GRPCServer{
		grpcServer: srv,
	}
}

// ShortRequest handles the request to save a URL and generates a short URL.
//
// Takes a context.Context and a pb.SaveURLRequest as input parameters.
// Returns a pb.SaveURLResponse and an error.
func (app *App) ShortRequest(ctx context.Context, req *pb.SaveURLRequest) (*pb.SaveURLResponse, error) {
	logger := logger.Get()
	isUniqueError := false

	_, userID := app.Session.AddUserSession()
	shortURL, err := app.storage.GetShortURL(userID, req.Url)
	result := app.Config.GetBaseAddr() + "/" + shortURL
	if err != nil {
		if errors.Is(err, storage.ErrUniqueViolation) {
			isUniqueError = true
		} else {
			logger.Error("gRPC server ShortRequest: cannot get short URL", zap.Error(err))
			return nil, fmt.Errorf("%w", status.Error(codes.InvalidArgument, err.Error()))
		}
	}

	return &pb.SaveURLResponse{
		Result:        result,
		IsUniqueError: isUniqueError,
	}, nil
}
