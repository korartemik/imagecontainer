package image_container

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"tagestest/internal/config"
	"tagestest/internal/dto"
	"tagestest/internal/lib/metrics"

	siv1 "tagestest/gen/go"

	"google.golang.org/grpc"
)

type serverAPI struct {
	siv1.UnimplementedImageServiceServer
	container        ImageContainer
	fileChannel      chan struct{}
	listFilesChannel chan struct{}
}

func newServerApi(container ImageContainer, cfg *config.Config) *serverAPI {
	fileChannel := make(chan struct{}, cfg.Server.LimitConcurrentFileRequest)
	listFilesChannel := make(chan struct{}, cfg.Server.LimitConcurrentListingRequest)
	return &serverAPI{
		container:        container,
		fileChannel:      fileChannel,
		listFilesChannel: listFilesChannel,
	}
}

type ImageContainer interface {
	Get(name string) (*dto.Image, error)
	Insert(image *dto.Image) error
	Select() ([]dto.ImageListFormat, error)
	Close() error
}

func Register(gRPCServer *grpc.Server, container ImageContainer, cfg *config.Config) {
	siv1.RegisterImageServiceServer(gRPCServer, newServerApi(container, cfg))
}

func (s *serverAPI) UploadImage(ctx context.Context, in *siv1.UploadImageRequest) (*siv1.UploadImageResponse, error) {
	metrics.GetProm().IncRequest(metrics.UploadMethod)

	select {
	case s.fileChannel <- struct{}{}:
	default:
		return nil, status.Error(codes.FailedPrecondition, "too many concurrent requests")
	}

	defer func() { <-s.fileChannel }()

	if in.Image.Filename == "" {
		return nil, status.Error(codes.InvalidArgument, "filename is required")
	}

	err := s.container.Insert(dto.ConvertImageToDto(in.Image))
	if err != nil {
		log.Error().Err(err).Msg("can not save image")
		return nil, status.Errorf(codes.Internal, "failed to save")
	}
	return &siv1.UploadImageResponse{Filename: in.Image.Filename}, nil
}

func (s *serverAPI) DownloadImage(ctx context.Context, in *siv1.DownloadImageRequest) (*siv1.DownloadImageResponse, error) {
	metrics.GetProm().IncRequest(metrics.GetMethod)
	select {
	case s.fileChannel <- struct{}{}:
	default:
		return nil, status.Error(codes.FailedPrecondition, "too many concurrent requests")
	}

	defer func() { <-s.fileChannel }()

	if in.Filename == "" {
		return nil, status.Error(codes.InvalidArgument, "filename is required")
	}

	image, err := s.container.Get(in.Filename)
	if err != nil {
		log.Error().Err(err).Msg("can not get image")
		return nil, status.Errorf(codes.Internal, "failed to download")
	}
	return &siv1.DownloadImageResponse{ImageData: dto.ConvertImageFromDto(image).ImageData}, nil
}

func (s *serverAPI) ListImages(ctx context.Context, in *siv1.ListImagesRequest) (*siv1.ListImagesResponse, error) {
	metrics.GetProm().IncRequest(metrics.ListMethod)

	select {
	case s.listFilesChannel <- struct{}{}:
	default:
		return nil, status.Error(codes.FailedPrecondition, "too many concurrent requests")
	}

	defer func() { <-s.listFilesChannel }()

	list, err := s.container.Select()
	if err != nil {
		log.Error().Err(err).Msg("can not list images")
		return nil, status.Errorf(codes.Internal, "failed to listing")
	}
	resp := &siv1.ListImagesResponse{Images: make([]*siv1.ImageNameFormat, len(list))}
	for _, format := range list {
		resp.Images = append(resp.Images, dto.ConvertImageListFormatFromDto(&format))
	}
	return resp, nil
}

func (s *serverAPI) Close() error {
	return s.container.Close()
}
