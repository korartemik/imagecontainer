package dto

import siv1 "tagestest/gen/go"

type Image struct {
	Name string
	Data []byte
}

func ConvertImageToDto(image *siv1.Image) *Image {
	return &Image{
		Name: image.Filename,
		Data: image.ImageData,
	}
}

func ConvertImageFromDto(image *Image) *siv1.Image {
	return &siv1.Image{
		Filename:  image.Name,
		ImageData: image.Data,
	}
}
