package model

import "time"

type Frame struct {
	ID                  int64             `json:"id"`
	VendorName          string            `json:"vendor_name"`
	FrameName           string            `json:"frame_name"`
	Color               string            `json:"color"`
	Description         string            `json:"description"`
	ArticleNumber       string            `json:"article_number"`
	ProductDetail       string            `json:"product_detail"`
	MaterialDescription string            `json:"material_description"`
	Care                string            `json:"care"`
	Price               float64           `json:"price"`
	Measurements        FrameMeasurements `json:"measurements"`
	Images              []FrameImage      `json:"images,omitempty"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

type FrameMeasurements struct {
	PictureWidthCm  float64 `json:"picture_width_cm"`
	PictureWidthIn  float64 `json:"picture_width_in"`
	PictureHeightCm float64 `json:"picture_height_cm"`
	PictureHeightIn float64 `json:"picture_height_in"`
	FrameWidthCm    float64 `json:"frame_width_cm"`
	FrameWidthIn    float64 `json:"frame_width_in"`
	FrameHeightCm   float64 `json:"frame_height_cm"`
	FrameHeightIn   float64 `json:"frame_height_in"`
	FrameDepthCm    float64 `json:"frame_depth_cm"`
	FrameDepthIn    float64 `json:"frame_depth_in"`
}

type FrameImage struct {
	ID                int64     `json:"id"`
	FrameID           int64     `json:"frame_id"`
	ImageURL          string    `json:"image_url"`
	ThumbnailURL      string    `json:"thumbnail_url"`
	OriginalURL       string    `json:"original_url"`
	BlobName          string    `json:"blob_name"`
	ThumbnailBlobName string    `json:"thumbnail_blob_name"`
	AltText           string    `json:"alt_text"`
	SortOrder         int       `json:"sort_order"`
	IsPrimary         bool      `json:"is_primary"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type SaveFrameRequest struct {
	VendorName          string            `json:"vendor_name"`
	FrameName           string            `json:"frame_name"`
	Color               string            `json:"color"`
	Description         string            `json:"description"`
	ArticleNumber       string            `json:"article_number"`
	ProductDetail       string            `json:"product_detail"`
	MaterialDescription string            `json:"material_description"`
	Care                string            `json:"care"`
	Price               float64           `json:"price"`
	Measurements        FrameMeasurements `json:"measurements"`
}

type SaveFrameImageRequest struct {
	BlobName  string `json:"blob_name"`
	AltText   string `json:"alt_text"`
	SortOrder int    `json:"sort_order"`
	IsPrimary bool   `json:"is_primary"`
}

type SaveFrameImagesRequest struct {
	MainImage     SaveFrameImageRequest   `json:"main_image"`
	GalleryImages []SaveFrameImageRequest `json:"gallery_images"`
	Images        []SaveFrameImageRequest `json:"images,omitempty"`
}
