package upload

import (
	"bytes"
	"fmt"
	"testing"
)

func TestUpload(t *testing.T) {
	fileContent := []byte("test")

	// 使用 bytes.Reader 代替 os.File
	sourceFile := bytes.NewReader(fileContent)

	filename, err := NewService().UploadFile(sourceFile, "./uploads", "test.txt")
	if err != nil {
		fmt.Println("failed to upload file, error:", err.Error())
		return
	}

	fmt.Println("filename", filename)
}

func TestDelete(t *testing.T) {
	filename := "uploads\\test.txt"
	err := NewService().DeleteFile(filename)
	if err != nil {
		fmt.Println("failed to delete file, error:", err.Error())
		return
	}
}
