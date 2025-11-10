package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

const UploadDir = "uploads/products" // Directory where images will be saved

// saveUploadedFile handles saving the file and returns the local path/URL.
func SaveUploadedFile(file *multipart.FileHeader, productID string) (string, error) {
	// 1. Ensure the upload directory exists
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// 2. Generate a unique filename using product ID and original extension
	extension := filepath.Ext(file.Filename)
	filename := productID + extension
	filePath := filepath.Join(UploadDir, filename)

	// 3. Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 4. Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// 5. Copy data to destination
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	// Return the relative path/URL
	return "/" + filePath, nil
}
