package types

import (
	"mime/multipart"
	"time"
)

// File represents a file uploaded by the user.
type File struct {
	Provider     string  `json:"provider"`
	Bucket       string  `json:"bucket"`
	FileType     string  `json:"fileType"`
	IsAnime      bool    `json:"isAnime"`
	TagName      string  `json:"tagName"`
	Landscape    string  `json:"landscape"`
	TopCategory  string  `json:"topCategory"`
	IsScreenshot bool    `json:"isScreenshot"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	ThumbW       float64 `json:"thumb_w"`
	ThumbH       float64 `json:"thumb_h"`
	ThumbSize    float64 `json:"thumb_size"`
	AlbumId      int64   `json:"albumId"`
}

// FileUploadMessage represents a message sent to the user after a file upload.
type FileUploadMessage struct {
	FaceID     int64                 `json:"face_id"`
	FileHeader *multipart.FileHeader `json:"fileHeader"`
	Result     File                  `json:"result"`
	UID        string                `json:"uid"`
	FilePath   string                `json:"filePath"`
	URL        string                `json:"url"`
	ThumbPath  string                `json:"thumbPath"`
	Thumbnail  string                `json:"thumbnail"`
}

type FileInfoResult struct {
	ID        int64     `json:"id"`
	FileName  string    `json:"file_name"`
	ThumbPath string    `json:"thumb_path"`
	ThumbW    float64   `json:"thumb_w"`
	ThumbH    float64   `json:"thumb_h"`
	ThumbSize float64   `json:"thumb_size"`
	CreatedAt time.Time `json:"created_at"`
	Path      string    `json:"path"`
}

type ThingImageList struct {
	ID        int64     `json:"id"`
	Category  string    `json:"category"`
	Tag       string    `json:"tag"`
	CreatedAt time.Time `json:"created_at"`
	ThumbPath string    `json:"thumb_path"`
	Path      string    `json:"path"`
}

type ShareImageInfo struct {
	Title          string `json:"title"`
	ExpireDate     string `json:"expire_date"`
	AccessLimit    int64  `json:"access_limit"`
	AccessPassword string `json:"access_password"`
	Provider       string `json:"provider"`
	Bucket         string `json:"bucket"`
}

type LocationInfo struct {
	ID         int64  `json:"id"`
	Country    string `json:"country"`
	City       string `json:"city"`
	Province   string `json:"province"`
	CoverImage string `json:"cover_image"`
	Total      int64  `json:"total"`
}
