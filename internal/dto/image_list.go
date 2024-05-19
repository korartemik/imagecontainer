package dto

import (
	siv1 "tagestest/gen/go"
	"time"
)

type ImageListFormat struct {
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ConvertImageListFormatFromDto(image *ImageListFormat) *siv1.ImageNameFormat {
	return &siv1.ImageNameFormat{
		Filename:  image.Name,
		CreatedAt: image.CreatedAt.UTC().String(),
		UpdatedAt: image.UpdatedAt.UTC().String(),
	}
}
