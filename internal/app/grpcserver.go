package app

import (
	"context"
	"encoding/json"
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
func (app *App) ShortRequest(ctx context.Context, req *pb.ShortRequestRequest) (*pb.ShortRequestResponse, error) {
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

	return &pb.ShortRequestResponse{
		Result:        result,
		IsUniqueError: isUniqueError,
	}, nil
}

func (app *App) ShortRequestBatch(ctx context.Context, req *pb.ShortRequestBatchRequest) (*pb.ShortRequestBatchResponse, error) {
	var rqJSON []storage.ReqJSONBatch

	logger := logger.Get()

	_, userID := app.Session.AddUserSession()

	body, err := json.Marshal(req.Items)
	if err != nil {
		logger.Error("gRPC server ShortRequestBatch: cannot marshal request body", zap.Error(err))
		return nil, fmt.Errorf("%w", status.Error(codes.InvalidArgument, err.Error()))
	}

	err = json.Unmarshal(body, &rqJSON)
	if err != nil {
		logger.Error("gRPC server ShortRequestBatch: cannot unmarshal request body", zap.Error(err))
		return nil, fmt.Errorf("%w", status.Error(codes.InvalidArgument, err.Error()))
	}

	rwJSON, err := app.storage.GetShortURLBatch(userID, app.Config.GetBaseAddr(), rqJSON)
	if err != nil {
		logger.Error("gRPC server ShortRequestBatch: cannot get short URL batch", zap.Error(err))
		return nil, fmt.Errorf("%w", status.Error(codes.InvalidArgument, err.Error()))
	}

	resItems := make([]*pb.ShortRequestBatchResponse_ShortRequestBatchItem, len(rwJSON))
	for i, item := range rwJSON {
		resItems[i] = &pb.ShortRequestBatchResponse_ShortRequestBatchItem{
			CorrelationId: item.ID,
			ShortUrl:      item.Result,
		}
	}

	return &pb.ShortRequestBatchResponse{
		Items: resItems,
	}, nil
}
