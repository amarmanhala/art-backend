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

func TestProductImageContentTypeValidation(t *testing.T) {
	allowed := []string{"image/jpeg", "image/png", "image/webp", "image/gif"}
	for _, contentType := range allowed {
		if !IsAllowedProductImageContentType(contentType) {
			t.Fatalf("expected %s to be allowed", contentType)
		}
	}

	if IsAllowedProductImageContentType("image/svg+xml") {
		t.Fatal("expected svg to be rejected")
	}
}

func TestCreateProductOriginalBlobName(t *testing.T) {
	blobName, err := CreateProductOriginalBlobName("art.PNG", "image/png")
	if err != nil {
		t.Fatalf("expected blob name, got error %v", err)
	}
	if !strings.HasPrefix(blobName, "products/originals/") {
		t.Fatalf("expected product original prefix, got %q", blobName)
	}
	if !strings.HasSuffix(blobName, ".png") {
		t.Fatalf("expected lower-case extension, got %q", blobName)
	}
}

func TestGenerateSquareJPEGThumbnail(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 800, 400))
	for y := 0; y < 400; y++ {
		for x := 0; x < 800; x++ {
			source.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}

	var input bytes.Buffer
	if err := jpeg.Encode(&input, source, nil); err != nil {
		t.Fatalf("could not encode source image: %v", err)
	}

	thumbnail, err := generateSquareJPEGThumbnail(input.Bytes(), config.DefaultProductThumbnailSize)
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
	if decoded.Bounds().Dx() != config.DefaultProductThumbnailSize || decoded.Bounds().Dy() != config.DefaultProductThumbnailSize {
		t.Fatalf("expected %dx%d thumbnail, got %dx%d", config.DefaultProductThumbnailSize, config.DefaultProductThumbnailSize, decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}
}

func TestProductImageServiceRejectsTooManyImages(t *testing.T) {
	service := NewProductImageService(nil, nil)
	request := model.SaveProductImagesRequest{Images: make([]model.SaveProductImageRequest, MaxProductImages+1)}

	_, err := service.ReplaceImages(context.Background(), 1, request)
	if err != ErrInvalidProductImages {
		t.Fatalf("expected ErrInvalidProductImages, got %v", err)
	}
}

func TestProductImageServiceRejectsNonProductOriginalBlob(t *testing.T) {
	service := NewProductImageService(nil, nil)
	request := model.SaveProductImagesRequest{
		MainImage: model.SaveProductImageRequest{BlobName: "carousel/example.jpg"},
	}

	_, err := service.ReplaceImages(context.Background(), 1, request)
	if err != ErrInvalidProductImages {
		t.Fatalf("expected ErrInvalidProductImages, got %v", err)
	}
}

func TestNormalizeProductImageRequestsUsesExplicitMainImage(t *testing.T) {
	request := model.SaveProductImagesRequest{
		MainImage: model.SaveProductImageRequest{
			BlobName: "products/originals/main.jpg",
		},
		GalleryImages: []model.SaveProductImageRequest{
			{BlobName: "products/originals/gallery.jpg"},
		},
	}

	images := normalizeProductImageRequests(request)
	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}
	if images[0].BlobName != "products/originals/main.jpg" || !images[0].IsPrimary || images[0].SortOrder != 1 {
		t.Fatalf("expected first image to be primary main image, got %+v", images[0])
	}
	if images[1].BlobName != "products/originals/gallery.jpg" || images[1].IsPrimary || images[1].SortOrder != 2 {
		t.Fatalf("expected second image to be non-primary gallery image, got %+v", images[1])
	}
}
