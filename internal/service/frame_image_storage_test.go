package service

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"testing"

	"art-backend/internal/config"
	"art-backend/internal/model"
)

func TestCreateFrameOriginalBlobName(t *testing.T) {
	blobName, err := CreateFrameOriginalBlobName("frame.PNG", "image/png")
	if err != nil {
		t.Fatalf("expected blob name, got error %v", err)
	}
	if !strings.HasPrefix(blobName, "frames/originals/") {
		t.Fatalf("expected frame original prefix, got %q", blobName)
	}
	if !strings.HasSuffix(blobName, ".png") {
		t.Fatalf("expected lower-case extension, got %q", blobName)
	}
}

func TestGenerateSquareJPEGThumbnailForFrames(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 800, 400))
	for y := 0; y < 400; y++ {
		for x := 0; x < 800; x++ {
			source.Set(x, y, color.RGBA{R: 40, G: 120, B: 180, A: 255})
		}
	}

	var input bytes.Buffer
	if err := jpeg.Encode(&input, source, nil); err != nil {
		t.Fatalf("could not encode source image: %v", err)
	}

	thumbnail, err := generateSquareJPEGThumbnail(input.Bytes(), config.DefaultFrameThumbnailSize)
	if err != nil {
		t.Fatalf("expected thumbnail, got error %v", err)
	}

	decoded, format, err := image.Decode(bytes.NewReader(thumbnail))
	if err != nil {
		t.Fatalf("could not decode thumbnail: %v", err)
	}
	if format != "jpeg" {
		t.Fatalf("expected jpeg thumbnail, got %q", format)
	}
	if decoded.Bounds().Dx() != config.DefaultFrameThumbnailSize || decoded.Bounds().Dy() != config.DefaultFrameThumbnailSize {
		t.Fatalf("expected %dx%d thumbnail, got %dx%d", config.DefaultFrameThumbnailSize, config.DefaultFrameThumbnailSize, decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}
}

func TestFrameImageServiceRejectsTooManyImages(t *testing.T) {
	service := NewFrameImageService(nil, nil)
	request := model.SaveFrameImagesRequest{Images: make([]model.SaveFrameImageRequest, MaxFrameImages+1)}

	_, err := service.ReplaceImages(context.Background(), 1, request)
	if err != ErrInvalidFrameImages {
		t.Fatalf("expected ErrInvalidFrameImages, got %v", err)
	}
}

func TestNormalizeFrameImageRequestsUsesExplicitMainImage(t *testing.T) {
	request := model.SaveFrameImagesRequest{
		MainImage: model.SaveFrameImageRequest{
			BlobName: "frames/originals/main.jpg",
		},
		GalleryImages: []model.SaveFrameImageRequest{
			{BlobName: "frames/originals/gallery.jpg"},
		},
	}

	images := normalizeFrameImageRequests(request)
	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}
	if images[0].BlobName != "frames/originals/main.jpg" || !images[0].IsPrimary || images[0].SortOrder != 1 {
		t.Fatalf("expected first image to be primary main image, got %+v", images[0])
	}
	if images[1].BlobName != "frames/originals/gallery.jpg" || images[1].IsPrimary || images[1].SortOrder != 2 {
		t.Fatalf("expected second image to be non-primary gallery image, got %+v", images[1])
	}
}
