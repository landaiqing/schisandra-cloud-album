package types

import "time"

type ShareFileInfoResult struct {
	ID        int64     `json:"id"`
	FileName  string    `json:"file_name"`
	ThumbPath string    `json:"thumb_path"`
	ThumbW    float64   `json:"thumb_w"`
	ThumbH    float64   `json:"thumb_h"`
	ThumbSize float64   `json:"thumb_size"`
	CreatedAt time.Time `json:"created_at"`
	Path      string    `json:"path"`
	Provider  string    `json:"provider"`
	Bucket    string    `json:"bucket"`
}
