package s3

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	apperrors "github.com/russo2642/renti_kz/pkg/errors"
)

type Storage struct {
	client     *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
	bucket     string
	region     string
}

func NewStorage(region, bucket string, accessKey, secretKey string) (*Storage, error) {
	if accessKey == "" || secretKey == "" {
		return nil, apperrors.NewStorageError("AWS credentials not set: accessKey and secretKey are required", nil)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(
			func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     accessKey,
					SecretAccessKey: secretKey,
				}, nil
			},
		)),
	)

	if err != nil {
		return nil, apperrors.NewStorageError("failed to load AWS config", err)
	}

	client := s3.NewFromConfig(cfg)

	uploader := manager.NewUploader(client)
	downloader := manager.NewDownloader(client)

	return &Storage{
		client:     client,
		uploader:   uploader,
		downloader: downloader,
		bucket:     bucket,
		region:     region,
	}, nil
}

func (s *Storage) UploadUserDocument(phone, docType string, data []byte) (string, error) {
	date := time.Now().Format("2006-01-02")
	fileExt := "jpg"
	objectKey := phone + "/doc/" + date + "/" + uuid.NewString() + "." + fileExt

	return s.uploadFile(objectKey, data)
}

func (s *Storage) UploadUserPhotoWithDoc(phone string, data []byte) (string, error) {
	date := time.Now().Format("2006-01-02")
	fileExt := "jpg"
	objectKey := phone + "/photo_with_doc/" + date + "/" + uuid.NewString() + "." + fileExt

	return s.uploadFile(objectKey, data)
}

func (s *Storage) UploadPropertyDocuments(phone string, data []byte, fileExt string) (string, error) {
	date := time.Now().Format("2006-01-02")
	if fileExt == "" {
		fileExt = "jpg"
	}
	objectKey := phone + "/kv_docs/" + date + "/" + uuid.NewString() + "." + fileExt

	return s.uploadFile(objectKey, data)
}

func (s *Storage) UploadFile(objectKey string, data []byte) (string, error) {
	return s.uploadFile(objectKey, data)
}

func (s *Storage) uploadFile(objectKey string, data []byte) (string, error) {
	reader := bytes.NewReader(data)

	_, err := s.uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
		Body:   reader,
	})

	if err != nil {
		return "", apperrors.NewStorageError("failed to upload file to S3", err)
	}

	fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, objectKey)
	return fileURL, nil
}

func (s *Storage) UploadMultipleFiles(files []FileUpload) ([]string, error) {
	if len(files) == 0 {
		return []string{}, nil
	}

	type uploadResult struct {
		index int
		url   string
		err   error
	}

	resultChan := make(chan uploadResult, len(files))

	for i, file := range files {
		go func(index int, upload FileUpload) {
			url, err := s.uploadFile(upload.ObjectKey, upload.Data)
			resultChan <- uploadResult{
				index: index,
				url:   url,
				err:   err,
			}
		}(i, file)
	}

	results := make([]string, len(files))
	var errors []error

	for i := 0; i < len(files); i++ {
		result := <-resultChan
		if result.err != nil {
			errors = append(errors, fmt.Errorf("файл %d: %w", result.index, result.err))
		} else {
			results[result.index] = result.url
		}
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("ошибки при загрузке файлов: %v", errors)
	}

	return results, nil
}

func (s *Storage) UploadUserDocumentsParallel(phone, docType string, documentsData [][]byte) ([]string, error) {
	date := time.Now().Format("2006-01-02")
	fileExt := "jpg"

	files := make([]FileUpload, len(documentsData))
	for i, data := range documentsData {
		objectKey := phone + "/doc/" + date + "/" + uuid.NewString() + "." + fileExt
		files[i] = FileUpload{
			ObjectKey: objectKey,
			Data:      data,
		}
	}

	return s.UploadMultipleFiles(files)
}

func (s *Storage) UploadApartmentPhotosParallel(apartmentID int, photosData [][]byte) ([]string, error) {
	date := time.Now().Format("2006-01-02")
	fileExt := "jpg"

	files := make([]FileUpload, len(photosData))
	for i, data := range photosData {
		objectKey := fmt.Sprintf("apartments/%d/photos/%s/%s.%s", apartmentID, date, uuid.NewString(), fileExt)
		files[i] = FileUpload{
			ObjectKey: objectKey,
			Data:      data,
		}
	}

	return s.UploadMultipleFiles(files)
}

type FileUpload struct {
	ObjectKey string
	Data      []byte
}

func (s *Storage) GetFileURL(objectKey string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, objectKey)
}

func (s *Storage) DeleteFile(objectKey string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		return apperrors.NewStorageError("failed to delete file from S3", err)
	}

	return nil
}

func (s *Storage) GetBaseURL() string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", s.bucket, s.region)
}

func (s *Storage) ExtractObjectKey(fullURL string) string {
	baseURL := s.GetBaseURL()
	if strings.HasPrefix(fullURL, baseURL) {
		return fullURL[len(baseURL):]
	}
	return fullURL
}
