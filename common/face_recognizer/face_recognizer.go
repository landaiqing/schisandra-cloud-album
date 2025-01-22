package face_recognizer

import (
	"github.com/Kagami/go-face"
	"os"
	"path/filepath"
)

// NewFaceRecognition creates a new instance of FaceRecognition
func NewFaceRecognition() *face.Recognizer {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
		return nil
	}
	modelDir := filepath.Join(dir, "/resources/models/face")
	rec, err := face.NewRecognizer(modelDir)
	if err != nil {
		panic(err)
		return nil
	}
	//defer rec.Close()
	return rec
}
