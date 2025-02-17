package file_upload_service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"sme-backend/src/core/config"
	"strings"
	"time"

	storage_go "github.com/supabase-community/storage-go"
)

type UploadedFileModel struct {
	FileName    string `json:"file_name"`
	Url         string `json:"url"`
	StoragePath string `json:"storage_path"`
}

func supabaseStorage() (*storage_go.Client, string) {
	url := config.Config("SUPABASE_URL")
	serviceKey := config.Config("SUPABASE_SERVICE_KEY")
	storageClient := storage_go.NewClient(url+"/storage/v1", serviceKey, nil)
	return storageClient, config.Config("PUBLIC_STORAGE_BUCKET_NAME")
}

func UploadFiles(response http.ResponseWriter, request *http.Request, storage_path string) ([]UploadedFileModel, error) {
	maxFileSize := int64(1024 * 1024 * 10) // 10MB
	if err := request.ParseMultipartForm(maxFileSize); err != nil {
		return nil, err
	}

	files := request.MultipartForm.File["files"]

	if len(files) == 0 {
		err := errors.New("files should not be empty")
		return nil, err
	}

	if storage_path == "" {
		err := errors.New("storage path is required")
		return nil, err
	}

	// Check each file size. File size should be less than 5MB
	for _, _file := range files {
		fileSize := _file.Size
		isFileSizeExceeded := fileSize > maxFileSize
		if isFileSizeExceeded {
			err := errors.New("file size should be less than 10MB")
			return nil, err
		}
	}

	uploaded_files := []UploadedFileModel{}
	for _, _file := range files {
		current_time_micro_unix := time.Now().UnixMicro()
		fileExtension := path.Ext(_file.Filename)
		complete_storage_path := fmt.Sprintf("%s/%d%s", storage_path, current_time_micro_unix, fileExtension)
		result, err := uploadToSupabaseStorage(_file, complete_storage_path)

		if err == nil {
			uploaded_files = append(uploaded_files, *result)
		}
	}

	return uploaded_files, nil
}

func uploadToSupabaseStorage(file *multipart.FileHeader, storagePath string) (*UploadedFileModel, error) {
	storageClient, bucketName := supabaseStorage()
	fileBody, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer fileBody.Close()

	bytes, err := io.ReadAll(fileBody)
	if err != nil {
		return nil, err
	}
	fileBody.Seek(0, io.SeekStart)
	contentType := http.DetectContentType(bytes)

	// Upload file to Supabase Storage
	_, err = storageClient.UploadFile(bucketName, storagePath, fileBody, storage_go.FileOptions{ContentType: &contentType})
	if err != nil {
		return nil, err
	}

	// Get public URL for the uploaded file
	response := storageClient.GetPublicUrl(bucketName, storagePath)
	fileUrl := response.SignedURL

	uploaded_file := &UploadedFileModel{}
	uploaded_file.FileName = file.Filename
	uploaded_file.Url = fileUrl
	uploaded_file.StoragePath = storagePath
	return uploaded_file, nil
}

func DeleteFileFromStorage(storagePaths []string) error {
	storageClient, bucketName := supabaseStorage()
	_, err := storageClient.RemoveFile(bucketName, storagePaths)
	return err
}

func GetStoragePathByURL(url string) string {
	var path string
	splitted := strings.Split(url, fmt.Sprintf("/%s/", config.Config("PUBLIC_STORAGE_BUCKET_NAME")))
	if len(splitted) == 2 {
		path = fmt.Sprintf("%s/%s", splitted[1], config.Config("PUBLIC_STORAGE_BUCKET_NAME"))
	}
	return path
}
