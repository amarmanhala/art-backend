package controller

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"art-backend/internal/config"
	"art-backend/internal/response"
	"art-backend/internal/service"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

type UploadController struct {
	config config.Config
}

type CreateCarouselUploadSASRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

type CreateCarouselUploadSASResponse struct {
	UploadURL string `json:"upload_url"`
	BlobName  string `json:"blob_name"`
	ImageURL  string `json:"image_url"`
	ExpiresAt string `json:"expires_at"`
}

type CreateProductImagesUploadSASRequest struct {
	Files []CreateProductImageUploadSASFile `json:"files"`
}

type CreateProductImageUploadSASFile struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

type CreateProductImagesUploadSASResponse struct {
	Uploads []CreateProductImageUploadSASResponse `json:"uploads"`
}

type CreateProductImageUploadSASResponse struct {
	UploadURL string `json:"upload_url"`
	BlobName  string `json:"blob_name"`
	ImageURL  string `json:"image_url"`
	ExpiresAt string `json:"expires_at"`
}

type CreateFrameImagesUploadSASRequest struct {
	Files []CreateFrameImageUploadSASFile `json:"files"`
}

type CreateFrameImageUploadSASFile struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

type CreateFrameImagesUploadSASResponse struct {
	Uploads []CreateFrameImageUploadSASResponse `json:"uploads"`
}

type CreateFrameImageUploadSASResponse struct {
	UploadURL string `json:"upload_url"`
	BlobName  string `json:"blob_name"`
	ImageURL  string `json:"image_url"`
	ExpiresAt string `json:"expires_at"`
}

type CreateArtStyleUploadSASRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

type CreateArtStyleUploadSASResponse struct {
	UploadURL string `json:"upload_url"`
	BlobName  string `json:"blob_name"`
	ImageURL  string `json:"image_url"`
	ExpiresAt string `json:"expires_at"`
}

func NewUploadController(config config.Config) *UploadController {
	return &UploadController{config: config}
}

func (c *UploadController) CreateCarouselUploadSAS(w http.ResponseWriter, r *http.Request) {
	if c.config.AzureStorageAccountName == "" ||
		c.config.AzureStorageAccountKey == "" ||
		c.config.AzureStorageContainer == "" {
		response.Error(w, http.StatusInternalServerError, "AZURE_NOT_CONFIGURED", "azure storage is not configured", "")
		return
	}

	var request CreateCarouselUploadSASRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	contentType := strings.TrimSpace(request.ContentType)
	if !isAllowedImageContentType(contentType) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid file type", "only jpeg, png, webp, and gif images are allowed")
		return
	}

	blobName, err := createCarouselBlobName(request.FileName, contentType)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create blob name", err.Error())
		return
	}

	credential, err := azblob.NewSharedKeyCredential(c.config.AzureStorageAccountName, c.config.AzureStorageAccountKey)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create azure credential", err.Error())
		return
	}

	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	sasQuery, err := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		ExpiryTime:    expiresAt,
		Permissions:   (&sas.BlobPermissions{Create: true, Write: true}).String(),
		ContainerName: c.config.AzureStorageContainer,
		BlobName:      blobName,
	}.SignWithSharedKey(credential)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create upload sas", err.Error())
		return
	}

	imageURL := c.blobURL(blobName)
	response.JSON(w, http.StatusOK, "carousel upload sas created successfully", CreateCarouselUploadSASResponse{
		UploadURL: imageURL + "?" + sasQuery.Encode(),
		BlobName:  blobName,
		ImageURL:  imageURL,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func (c *UploadController) CreateProductImagesUploadSAS(w http.ResponseWriter, r *http.Request) {
	if c.config.AzureStorageAccountName == "" ||
		c.config.AzureStorageAccountKey == "" ||
		c.config.AzureProductImagesContainer == "" {
		response.Error(w, http.StatusInternalServerError, "AZURE_NOT_CONFIGURED", "azure product image storage is not configured", "")
		return
	}

	var request CreateProductImagesUploadSASRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}
	if len(request.Files) == 0 || len(request.Files) > service.MaxProductImages {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "send a files array with at least 1 and at most 10 image files")
		return
	}

	credential, err := azblob.NewSharedKeyCredential(c.config.AzureStorageAccountName, c.config.AzureStorageAccountKey)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create azure credential", err.Error())
		return
	}

	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	uploads := make([]CreateProductImageUploadSASResponse, 0, len(request.Files))
	for _, file := range request.Files {
		contentType := strings.TrimSpace(file.ContentType)
		if !service.IsAllowedProductImageContentType(contentType) {
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid file type", "only jpeg, png, webp, and gif images are allowed")
			return
		}

		blobName, err := service.CreateProductOriginalBlobName(file.FileName, contentType)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create blob name", err.Error())
			return
		}

		sasQuery, err := sas.BlobSignatureValues{
			Protocol:      sas.ProtocolHTTPS,
			ExpiryTime:    expiresAt,
			Permissions:   (&sas.BlobPermissions{Create: true, Write: true}).String(),
			ContainerName: c.config.AzureProductImagesContainer,
			BlobName:      blobName,
		}.SignWithSharedKey(credential)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create upload sas", err.Error())
			return
		}

		imageURL := c.productImageBlobURL(blobName)
		uploads = append(uploads, CreateProductImageUploadSASResponse{
			UploadURL: imageURL + "?" + sasQuery.Encode(),
			BlobName:  blobName,
			ImageURL:  imageURL,
			ExpiresAt: expiresAt.Format(time.RFC3339),
		})
	}

	response.JSON(w, http.StatusOK, "product image upload sas created successfully", CreateProductImagesUploadSASResponse{Uploads: uploads})
}

func (c *UploadController) CreateFrameImagesUploadSAS(w http.ResponseWriter, r *http.Request) {
	if c.config.AzureStorageAccountName == "" ||
		c.config.AzureStorageAccountKey == "" ||
		c.config.AzureFrameImagesContainer == "" {
		response.Error(w, http.StatusInternalServerError, "AZURE_NOT_CONFIGURED", "azure frame image storage is not configured", "")
		return
	}

	var request CreateFrameImagesUploadSASRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}
	if len(request.Files) == 0 || len(request.Files) > service.MaxFrameImages {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "send a files array with at least 1 and at most 10 image files")
		return
	}

	credential, err := azblob.NewSharedKeyCredential(c.config.AzureStorageAccountName, c.config.AzureStorageAccountKey)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create azure credential", err.Error())
		return
	}

	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	uploads := make([]CreateFrameImageUploadSASResponse, 0, len(request.Files))
	for _, file := range request.Files {
		contentType := strings.TrimSpace(file.ContentType)
		if !service.IsAllowedFrameImageContentType(contentType) {
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid file type", "only jpeg, png, webp, and gif images are allowed")
			return
		}

		blobName, err := service.CreateFrameOriginalBlobName(file.FileName, contentType)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create blob name", err.Error())
			return
		}

		sasQuery, err := sas.BlobSignatureValues{
			Protocol:      sas.ProtocolHTTPS,
			ExpiryTime:    expiresAt,
			Permissions:   (&sas.BlobPermissions{Create: true, Write: true}).String(),
			ContainerName: c.config.AzureFrameImagesContainer,
			BlobName:      blobName,
		}.SignWithSharedKey(credential)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create upload sas", err.Error())
			return
		}

		imageURL := c.frameImageBlobURL(blobName)
		uploads = append(uploads, CreateFrameImageUploadSASResponse{
			UploadURL: imageURL + "?" + sasQuery.Encode(),
			BlobName:  blobName,
			ImageURL:  imageURL,
			ExpiresAt: expiresAt.Format(time.RFC3339),
		})
	}

	response.JSON(w, http.StatusOK, "frame image upload sas created successfully", CreateFrameImagesUploadSASResponse{Uploads: uploads})
}

func (c *UploadController) CreateArtStyleUploadSAS(w http.ResponseWriter, r *http.Request) {
	if c.config.AzureStorageAccountName == "" ||
		c.config.AzureStorageAccountKey == "" ||
		c.config.AzureArtStylesContainer == "" {
		response.Error(w, http.StatusInternalServerError, "AZURE_NOT_CONFIGURED", "azure art style image storage is not configured", "")
		return
	}

	var request CreateArtStyleUploadSASRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	contentType := strings.TrimSpace(request.ContentType)
	if !isAllowedImageContentType(contentType) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid file type", "only jpeg, png, webp, and gif images are allowed")
		return
	}

	blobName, err := createArtStyleBlobName(request.FileName, contentType)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create blob name", err.Error())
		return
	}

	credential, err := azblob.NewSharedKeyCredential(c.config.AzureStorageAccountName, c.config.AzureStorageAccountKey)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create azure credential", err.Error())
		return
	}

	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	sasQuery, err := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		ExpiryTime:    expiresAt,
		Permissions:   (&sas.BlobPermissions{Create: true, Write: true}).String(),
		ContainerName: c.config.AzureArtStylesContainer,
		BlobName:      blobName,
	}.SignWithSharedKey(credential)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create upload sas", err.Error())
		return
	}

	imageURL := c.artStyleBlobURL(blobName)
	response.JSON(w, http.StatusOK, "art style upload sas created successfully", CreateArtStyleUploadSASResponse{
		UploadURL: imageURL + "?" + sasQuery.Encode(),
		BlobName:  blobName,
		ImageURL:  imageURL,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func (c *UploadController) blobURL(blobName string) string {
	baseURL := strings.TrimRight(c.config.AzureStoragePublicURL, "/")
	if baseURL == "" {
		baseURL = "https://" + c.config.AzureStorageAccountName + ".blob.core.windows.net/" + c.config.AzureStorageContainer
	}

	return baseURL + "/" + escapeBlobPath(blobName)
}

func (c *UploadController) productImageBlobURL(blobName string) string {
	baseURL := strings.TrimRight(c.config.AzureProductImagesPublicURL, "/")
	if baseURL == "" {
		baseURL = "https://" + c.config.AzureStorageAccountName + ".blob.core.windows.net/" + c.config.AzureProductImagesContainer
	}

	return baseURL + "/" + escapeBlobPath(blobName)
}

func (c *UploadController) frameImageBlobURL(blobName string) string {
	baseURL := strings.TrimRight(c.config.AzureFrameImagesPublicURL, "/")
	if baseURL == "" {
		baseURL = "https://" + c.config.AzureStorageAccountName + ".blob.core.windows.net/" + c.config.AzureFrameImagesContainer
	}

	return baseURL + "/" + escapeBlobPath(blobName)
}

func (c *UploadController) artStyleBlobURL(blobName string) string {
	baseURL := strings.TrimRight(c.config.AzureArtStylesPublicURL, "/")
	if baseURL == "" {
		baseURL = "https://" + c.config.AzureStorageAccountName + ".blob.core.windows.net/" + c.config.AzureArtStylesContainer
	}

	return baseURL + "/" + escapeBlobPath(blobName)
}

func createCarouselBlobName(fileName string, contentType string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	extension := strings.ToLower(path.Ext(fileName))
	if extension == "" {
		extension = extensionFromContentType(contentType)
	}

	return "carousel/" + hex.EncodeToString(randomBytes) + extension, nil
}

func isAllowedImageContentType(contentType string) bool {
	switch strings.ToLower(contentType) {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return true
	default:
		return false
	}
}

func extensionFromContentType(contentType string) string {
	switch strings.ToLower(contentType) {
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

func createArtStyleBlobName(fileName string, contentType string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	extension := strings.ToLower(path.Ext(fileName))
	if extension == "" {
		extension = extensionFromContentType(contentType)
	}

	return "art-styles/" + hex.EncodeToString(randomBytes) + extension, nil
}
