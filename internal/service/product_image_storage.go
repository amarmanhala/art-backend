package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/url"
	"path"
	"strings"

	"art-backend/internal/config"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

var ErrProductImageStorageNotConfigured = errors.New("product image storage is not configured")

type ProductImageStorage interface {
	CreateThumbnail(ctx context.Context, originalBlobName string) (ProductImageAsset, error)
	ImageURL(blobName string) string
}

type ProductImageAsset struct {
	ImageURL          string
	OriginalURL       string
	ThumbnailURL      string
	BlobName          string
	ThumbnailBlobName string
}

type AzureProductImageStorage struct {
	accountName string
	accountKey  string
	container   string
	publicURL   string
	thumbSize   int
}

func NewAzureProductImageStorage(config config.Config) *AzureProductImageStorage {
	return &AzureProductImageStorage{
		accountName: strings.TrimSpace(config.AzureStorageAccountName),
		accountKey:  strings.TrimSpace(config.AzureStorageAccountKey),
		container:   strings.TrimSpace(config.AzureProductImagesContainer),
		publicURL:   strings.TrimRight(strings.TrimSpace(config.AzureProductImagesPublicURL), "/"),
		thumbSize:   config.ProductThumbnailSize,
	}
}

func (s *AzureProductImageStorage) CreateThumbnail(ctx context.Context, originalBlobName string) (ProductImageAsset, error) {
	if s.accountName == "" || s.accountKey == "" || s.container == "" {
		return ProductImageAsset{}, ErrProductImageStorageNotConfigured
	}

	client, err := s.client()
	if err != nil {
		return ProductImageAsset{}, err
	}

	download, err := client.DownloadStream(ctx, s.container, originalBlobName, nil)
	if err != nil {
		return ProductImageAsset{}, err
	}
	defer download.Body.Close()

	originalBytes, err := io.ReadAll(download.Body)
	if err != nil {
		return ProductImageAsset{}, err
	}

	thumbnailSize := s.thumbSize
	if thumbnailSize <= 0 {
		thumbnailSize = config.DefaultProductThumbnailSize
	}

	thumbnail, err := generateSquareJPEGThumbnail(originalBytes, thumbnailSize)
	if err != nil {
		return ProductImageAsset{}, err
	}

	thumbnailBlobName, err := createRandomBlobName("products/thumbnails", ".jpg")
	if err != nil {
		return ProductImageAsset{}, err
	}

	contentType := "image/jpeg"
	_, err = client.UploadBuffer(ctx, s.container, thumbnailBlobName, thumbnail, &azblob.UploadBufferOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType:  &contentType,
			BlobCacheControl: to.Ptr("public, max-age=31536000, immutable"),
		},
	})
	if err != nil {
		return ProductImageAsset{}, err
	}

	originalURL := s.ImageURL(originalBlobName)
	thumbnailURL := s.ImageURL(thumbnailBlobName)
	return ProductImageAsset{
		ImageURL:          originalURL,
		OriginalURL:       originalURL,
		ThumbnailURL:      thumbnailURL,
		BlobName:          originalBlobName,
		ThumbnailBlobName: thumbnailBlobName,
	}, nil
}

func (s *AzureProductImageStorage) ImageURL(blobName string) string {
	baseURL := s.publicURL
	if baseURL == "" {
		baseURL = "https://" + s.accountName + ".blob.core.windows.net/" + s.container
	}

	return baseURL + "/" + escapeBlobPath(blobName)
}

func (s *AzureProductImageStorage) client() (*azblob.Client, error) {
	credential, err := azblob.NewSharedKeyCredential(s.accountName, s.accountKey)
	if err != nil {
		return nil, err
	}

	return azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", s.accountName), credential, nil)
}

func generateSquareJPEGThumbnail(data []byte, size int) ([]byte, error) {
	source, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		source, err = webp.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	}

	bounds := source.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, errors.New("invalid image dimensions")
	}

	cropSize := width
	if height < cropSize {
		cropSize = height
	}
	cropX := bounds.Min.X + (width-cropSize)/2
	cropY := bounds.Min.Y + (height-cropSize)/2
	crop := image.Rect(cropX, cropY, cropX+cropSize, cropY+cropSize)

	destination := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(destination, destination.Bounds(), source, crop, draw.Over, nil)

	var output bytes.Buffer
	if err := jpeg.Encode(&output, destination, &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

func CreateProductOriginalBlobName(fileName string, contentType string) (string, error) {
	extension := strings.ToLower(path.Ext(fileName))
	if extension == "" {
		extension = extensionFromImageContentType(contentType)
	}
	if extension == "" {
		return "", errors.New("unsupported file type")
	}

	return createRandomBlobName("products/originals", extension)
}

func createRandomBlobName(prefix string, extension string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return strings.TrimRight(prefix, "/") + "/" + hex.EncodeToString(randomBytes) + extension, nil
}

func IsAllowedProductImageContentType(contentType string) bool {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return true
	default:
		return false
	}
}

func extensionFromImageContentType(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}

func escapeBlobPath(blobName string) string {
	parts := strings.Split(blobName, "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(part)
	}

	return strings.Join(parts, "/")
}
