package types

// File represents a file uploaded by the user.
type File struct {
	Provider     string   `json:"provider"`
	Bucket       string   `json:"bucket"`
	FileType     string   `json:"fileType"`
	IsAnime      bool     `json:"isAnime"`
	ObjectArray  []string `json:"objectArray"`
	Landscape    string   `json:"landscape"`
	TopCategory  string   `json:"topCategory"`
	IsScreenshot bool     `json:"isScreenshot"`
	Exif         any      `json:"exif"`
}
