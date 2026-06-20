package service

import (
	"context"
	"errors"
	"fmt"
	_ "image/jpeg"
	"io"
	"path"
	"strings"

	"art-backend/internal/config"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

var ErrFrameImageStorageNotConfigured = errors.New("frame image storage is not configured")

type FrameImageStorage interface {
	CreateThumbnail(ctx context.Context, originalBlobName string) (FrameImageAsset, error)
	ImageURL(blobName string) string
}

type FrameImageAsset struct {
	ImageURL          string
	OriginalURL       string
	ThumbnailURL      string
	BlobName          string
	ThumbnailBlobName string
}

type AzureFrameImageStorage struct {
	accountName string
	accountKey  string
	container   string
	publicURL   string
	thumbSize   int
}

func NewAzureFrameImageStorage(config config.Config) *AzureFrameImageStorage {
	return &AzureFrameImageStorage{
		accountName: strings.TrimSpace(config.AzureStorageAccountName),
		accountKey:  strings.TrimSpace(config.AzureStorageAccountKey),
		container:   strings.TrimSpace(config.AzureFrameImagesContainer),
		publicURL:   strings.TrimRight(strings.TrimSpace(config.AzureFrameImagesPublicURL), "/"),
		thumbSize:   config.FrameThumbnailSize,
	}
}

func (s *AzureFrameImageStorage) CreateThumbnail(ctx context.Context, originalBlobName string) (FrameImageAsset, error) {
	if s.accountName == "" || s.accountKey == "" || s.container == "" {
		return FrameImageAsset{}, ErrFrameImageStorageNotConfigured
	}

	client, err := s.client()
	if err != nil {
		return FrameImageAsset{}, err
	}

	download, err := client.DownloadStream(ctx, s.container, originalBlobName, nil)
	if err != nil {
		return FrameImageAsset{}, err
	}
	defer download.Body.Close()

	originalBytes, err := io.ReadAll(download.Body)
	if err != nil {
		return FrameImageAsset{}, err
	}

	thumbnailSize := s.thumbSize
	if thumbnailSize <= 0 {
		thumbnailSize = config.DefaultFrameThumbnailSize
	}

	thumbnail, err := generateSquareJPEGThumbnail(originalBytes, thumbnailSize)
	if err != nil {
		return FrameImageAsset{}, err
	}

	thumbnailBlobName, err := createRandomBlobName("frames/thumbnails", ".jpg")
	if err != nil {
		return FrameImageAsset{}, err
	}

	contentType := "image/jpeg"
	_, err = client.UploadBuffer(ctx, s.container, thumbnailBlobName, thumbnail, &azblob.UploadBufferOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType:  &contentType,
			BlobCacheControl: to.Ptr("public, max-age=31536000, immutable"),
		},
	})
	if err != nil {
		return FrameImageAsset{}, err
	}

	originalURL := s.ImageURL(originalBlobName)
	thumbnailURL := s.ImageURL(thumbnailBlobName)
	return FrameImageAsset{
		ImageURL:          originalURL,
		OriginalURL:       originalURL,
		ThumbnailURL:      thumbnailURL,
		BlobName:          originalBlobName,
		ThumbnailBlobName: thumbnailBlobName,
	}, nil
}

func (s *AzureFrameImageStorage) ImageURL(blobName string) string {
	baseURL := s.publicURL
	if baseURL == "" {
		baseURL = "https://" + s.accountName + ".blob.core.windows.net/" + s.container
	}

	return baseURL + "/" + escapeBlobPath(blobName)
}

func (s *AzureFrameImageStorage) client() (*azblob.Client, error) {
	credential, err := azblob.NewSharedKeyCredential(s.accountName, s.accountKey)
	if err != nil {
		return nil, err
	}

	return azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", s.accountName), credential, nil)
}

func CreateFrameOriginalBlobName(fileName string, contentType string) (string, error) {
	extension := strings.ToLower(path.Ext(fileName))
	if extension == "" {
		extension = extensionFromImageContentType(contentType)
	}
	if extension == "" {
		return "", errors.New("unsupported file type")
	}

	return createRandomBlobName("frames/originals", extension)
}

func IsAllowedFrameImageContentType(contentType string) bool {
	return IsAllowedProductImageContentType(contentType)
}
