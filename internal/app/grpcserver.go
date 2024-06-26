package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/stsg/shorty/api/v1"
	"github.com/stsg/shorty/internal/logger"
	"github.com/stsg/shorty/internal/storage"
)

// GRPCServer is a struct that holds gRPC server data.
type GRPCServer struct {
	pb.UnimplementedShortenerServiceServer
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

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(interceptors...)),
	)
	pb.RegisterShortenerServiceServer(srv, &GRPCServer{
		grpcServer: srv,
	})

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

	_, userID := app.Session.AddUserSession()
	shortURL, err := app.storage.GetShortURL(userID, req.Url)
	result := app.Config.GetBaseAddr() + "/" + shortURL
	if err != nil {
		if errors.Is(err, storage.ErrUniqueViolation) {
			return &pb.ShortRequestResponse{
				Result: result,
				Error:  status.Error(codes.AlreadyExists, err.Error()).Error(),
			}, nil
		} else {
			logger.Error("gRPC server ShortRequest: cannot get short URL", zap.Error(err))
			return &pb.ShortRequestResponse{
				Result: result,
				Error:  status.Error(codes.InvalidArgument, err.Error()).Error(),
			}, fmt.Errorf("%w", status.Error(codes.InvalidArgument, err.Error()))
		}
	}

	return &pb.ShortRequestResponse{
		Result: result,
		Error:  status.Error(codes.OK, "").Error(),
	}, nil
}

// ShortID retrieves the long URL associated with the given short ID.
//
// ctx: The context for the function.
// req: The ShortIDRequest containing the short ID.
// Returns the ShortIDResponse containing the long URL and a flag indicating if the URL is deleted.
// Returns an error if the long URL cannot be retrieved.
func (app *App) ShortID(ctx context.Context, req *pb.ShortIDRequest) (*pb.ShortIDResponse, error) {
	logger := logger.Get()

	id := strings.TrimPrefix(req.Url, "/")
	id = strings.TrimSuffix(id, "/")
	longURL, err := app.storage.GetRealURL(id)
	if err != nil {
		if errors.Is(err, storage.ErrURLDeleted) {
			return &pb.ShortIDResponse{
				Result: longURL,
				Error:  status.Error(codes.NotFound, err.Error()).Error(),
			}, nil
		} else {
			logger.Error("gRPC server ShortRequest: cannot get long URL", zap.Error(err))
			return &pb.ShortIDResponse{
				Result: longURL,
				Error:  status.Error(codes.InvalidArgument, err.Error()).Error(),
			}, fmt.Errorf("%w", status.Error(codes.InvalidArgument, err.Error()))
		}
	}

	return &pb.ShortIDResponse{
		Result: longURL,
		Error:  status.Error(codes.OK, "").Error(),
	}, nil
}

// ShortRequestBatch handles the request to save a batch of URLs and generates short URLs.
//
// Takes a context.Context and a pb.ShortRequestBatchRequest as input parameters.
// Returns a pb.ShortRequestBatchResponse and an error.
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

// GetStats retrieves the statistics of URLs and users from the App's storage.
//
// It takes a context and an empty protobuf message as input parameters.
// It returns a GetStatsResponse protobuf message containing the number of URLs and users, and an error if any.
func (app *App) GetStats(ctx context.Context, _ *emptypb.Empty) (*pb.GetStatsResponse, error) {
	logger := logger.Get()

	stats, err := app.storage.GetStats()
	if err != nil {
		logger.Error("gRPC server GetStats: cannot get stats", zap.Error(err))
		return nil, fmt.Errorf("%w", status.Error(codes.InvalidArgument, err.Error()))
	}

	return &pb.GetStatsResponse{
		Urls:  uint32(stats.URLCount),
		Users: uint32(stats.UserCount),
	}, nil
}
