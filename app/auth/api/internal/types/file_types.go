package types

import (
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
	HasQrcode    bool    `json:"hasQrcode"`
}

type UploadSetting struct {
	NsfwDetection       bool `json:"nsfw_detection"`
	AnimeDetection      bool `json:"anime_detection"`
	LandscapeDetection  bool `json:"landscape_detection"`
	ScreenshotDetection bool `json:"screenshot_detection"`
	GpsDetection        bool `json:"gps_detection"`
	TargetDetection     bool `json:"target_detection"`
	QrcodeDetection     bool `json:"qrcode_detection"`
	FaceDetection       bool `json:"face_detection"`
	Encrypt             bool `json:"encrypt"`
}

type FileUploadMessage struct {
	FaceID    int64         `json:"face_id"`
	FileName  string        `json:"file_name"`
	FileSize  int64         `json:"file_size"`
	Result    File          `json:"result"`
	UID       string        `json:"uid"`
	FilePath  string        `json:"file_path"`
	ThumbPath string        `json:"thumb_path"`
	Setting   UploadSetting `json:"setting"`
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

type ZincFileInfo struct {
	FaceID       int64     `json:"face_id"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	UID          string    `json:"uid"`
	FilePath     string    `json:"file_path"`
	ThumbPath    string    `json:"thumb_path"`
	CreatedAt    time.Time `json:"created_at"`
	StorageId    int64     `json:"storage_id"`
	Provider     string    `json:"provider"`
	Bucket       string    `json:"bucket"`
	FileType     string    `json:"file_type"`
	IsAnime      bool      `json:"is_anime"`
	TagName      string    `json:"tag_name"`
	Landscape    string    `json:"landscape"`
	TopCategory  string    `json:"top_category"`
	IsScreenshot bool      `json:"is_screenshot"`
	Width        float64   `json:"width"`
	Height       float64   `json:"height"`
	Longitude    float64   `json:"longitude"`
	Latitude     float64   `json:"latitude"`
	ThumbW       float64   `json:"thumb_w"`
	ThumbH       float64   `json:"thumb_h"`
	ThumbSize    float64   `json:"thumb_size"`
	AlbumId      int64     `json:"album_id"`
	HasQrcode    bool      `json:"has_qrcode"`
	Country      string    `json:"country"`
	Province     string    `json:"province"`
	City         string    `json:"city"`
}

type ImageBedMeta struct {
	Provider string `json:"provider"`
	Bucket   string `json:"bucket"`
	Width    int64  `json:"width"`
	Height   int64  `json:"height"`
}
