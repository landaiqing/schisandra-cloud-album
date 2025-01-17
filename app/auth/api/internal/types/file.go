package types

// File represents a file uploaded by the user.
type File struct {
	UID             string `json:"uid"`
	FileName        string `json:"fileName"`
	FileType        string `json:"fileType"`
	DetectionResult struct {
		IsAnime     bool     `json:"isAnime"`
		HasFace     bool     `json:"hasFace"`
		ObjectArray []string `json:"objectArray"`
		Landscape   string   `json:"landscape"`
	}
}
