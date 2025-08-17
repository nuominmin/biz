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

	svc := NewService("http://127.0.0.1:3000", "goods")
	filename, err := svc.UploadFile(sourceFile, "test.txt")
	if err != nil {
		fmt.Println("failed to upload file, error:", err.Error())
		return
	}

	fmt.Println("filename", filename)
}

func TestDelete(t *testing.T) {
	filename := "uploads\\goods\\test.txt"
	svc := NewService("http://127.0.0.1:3000", "goods")
	err := svc.DeleteFile(filename)
	if err != nil {
		fmt.Println("failed to delete file, error:", err.Error())
		return
	}
}
