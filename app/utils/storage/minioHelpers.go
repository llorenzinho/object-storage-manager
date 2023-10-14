package storage

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	dto "github.com/llorenzinho/object-storage-manager/models/DTO"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/tags"
)

// MinioStorage is a struct handling the connection to the Minio storage
type MinioStorageHelper struct {
	config *StorageConfig
	client *minio.Client
}

// MinioStorageHelper singleton
var minioStorageHelper *MinioStorageHelper = nil

// Ping minio storage
func (m *MinioStorageHelper) Ping() error {
	// Timeout context of 10 seconds.
	cancel, err := m.client.HealthCheck(time.Second * 10)
	defer cancel()
	if err != nil {
		return err
	}
	return nil
}

// NewMinioStorage returns a new MinioStorage singleton
func Instance() *MinioStorageHelper {
	if minioStorageHelper == nil {
		// Create the minio client
		conf := ReadConfig()
		opts := &minio.Options{
			Creds: credentials.NewStaticV4(
				conf.MinioStorageConfigStruct.MinioStorageConfig.Auth.AccessKeyID,
				conf.MinioStorageConfigStruct.MinioStorageConfig.Auth.SecretAccessKey, ""),
			Secure: false,
		}
		client, err := minio.New(
			fmt.Sprintf("%s:%d", conf.MinioStorageConfigStruct.MinioStorageConfig.Url, conf.MinioStorageConfigStruct.MinioStorageConfig.Port),
			opts,
		)
		if err != nil {
			panic(fmt.Errorf("unable to create minio client: %w", err))
		}
		minioStorageHelper = &MinioStorageHelper{
			config: conf,
			client: client,
		}
		if minioStorageHelper.Ping() != nil {
			panic(fmt.Errorf("unable to ping minio storage: %w", err))
		}
		log.Default().Println("Minio storage initialized")
	}

	return minioStorageHelper
}

// Parse minio object into dto.File
// func parseMinioObject(obj *minio.Object, bucket string) (*dto.File, error) {
// 	defer obj.Close()
// 	fileStat, err := obj.Stat()
// 	if err != nil {
// 		return nil, err
// 	}
// 	f := parseMinioObjectInfo(fileStat, bucket)

// 	return f, nil
// }

// Parse minio ObjectInfo into dto.File
func parseMinioObjectInfo(obj minio.ObjectInfo, bucket string) *dto.File {
	f := &dto.File{
		Name:        obj.Key,
		Bucket:      bucket,
		Verified:    false,
		ContentType: obj.ContentType,
	}
	return f
}

// Enhance dto.File with tags
func enhanceFileWithTags(ctx context.Context, client *minio.Client, file *dto.File) error {
	tags, err := client.GetObjectTagging(ctx, file.Bucket, file.Name, minio.GetObjectTaggingOptions{})
	if err != nil {
		return err
	}
	verified, err := strconv.ParseBool(tags.ToMap()["verified"])
	if err != nil {
		return fmt.Errorf("unable to parse verified tag: %w", err)
	}
	file.Verified = verified
	return nil
}

// Get file from minio given the bucket name and the file name
// Returns the file as a dto.File
func (m *MinioStorageHelper) Get(bucket string, fileName string, ctx context.Context) (*dto.File, error) {
	obj, err := m.client.StatObject(ctx, bucket, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	file := parseMinioObjectInfo(obj, bucket)
	if err != nil {
		return nil, err
	}
	err = enhanceFileWithTags(ctx, m.client, file)
	if err != nil {
		_ = fmt.Errorf("unable to enhance file with tags: %w", err)
	}
	return file, nil
}

// Get all files from minio given the bucket name
// Returns the files as a slice of dto.File
func (m *MinioStorageHelper) GetAll(bucket string, ctx context.Context) (*dto.FileList, error) {
	files := dto.FileList{}
	doneCh := make(chan struct{})
	defer close(doneCh)
	for obj := range m.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Recursive: false}) {
		if obj.Err != nil {
			return nil, obj.Err
		}
		file := parseMinioObjectInfo(obj, bucket)
		err := enhanceFileWithTags(ctx, m.client, file)
		if err != nil {
			_ = fmt.Errorf("unable to enhance file with tags: %w", err)
		}
		files = append(files, *file)
	}
	return &files, nil
}

// Upload file to minio given the bucket name, the file name and the file path and the content bytes and type
// Returns the file as a dto.File
// Add tags to the file
func (m *MinioStorageHelper) Upload(
	bucket string, filePath string,
	content []byte, contentType string, ctx context.Context,
) (*dto.File, error) {
	if exists, _ := m.Exists(bucket, filePath, ctx); exists {
		return nil, NewFileAlreadyExistsError(bucket, filePath) // TODO: add error code
	}
	objSize := int64(len(content))
	objReader := bytes.NewReader(content)
	opts := minio.PutObjectOptions{ContentType: contentType, UserTags: map[string]string{"verified": "false"}}
	_, err := m.client.PutObject(ctx, bucket, filePath,
		objReader, objSize, opts)
	if err != nil {
		return nil, err
	}
	file, err := m.Get(bucket, filePath, ctx)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Delete file from minio given the bucket name and the file name
// Returns the file as a dto.File
func (m *MinioStorageHelper) Delete(bucket string, filePath string, ctx context.Context) (*dto.File, error) {
	if exists, _ := m.Exists(bucket, filePath, ctx); !exists {
		return nil, NewFileDoesNotExistError(bucket, filePath) // TODO: add error code
	}

	file, err := m.Get(bucket, filePath, ctx)
	if err != nil {
		return nil, err
	}
	err = m.client.RemoveObject(ctx, bucket, filePath, minio.RemoveObjectOptions{})
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Set file tags from minio given the bucket name, the file name and the tags
// Returns the file as a dto.File
func (m *MinioStorageHelper) setTags(bucket string, filePath string, key string, value string, ctx context.Context) (*dto.File, error) {
	if exists, _ := m.Exists(bucket, filePath, ctx); !exists {
		return nil, NewFileDoesNotExistError(bucket, filePath) // TODO: add error code
	}
	t, err := tags.MapToObjectTags(map[string]string{key: value})
	if err != nil {
		return nil, err
	}
	err = m.client.PutObjectTagging(ctx, bucket, filePath, t, minio.PutObjectTaggingOptions{})
	if err != nil {
		return nil, err
	}
	file, err := m.Get(bucket, filePath, ctx)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Verify file from minio given the bucket name and the file name
// Returns the file as a dto.File
func (m *MinioStorageHelper) Verify(bucket string, filePath string, ctx context.Context) (*dto.File, error) {
	return m.setTags(bucket, filePath, "verified", "true", ctx)
}

// Unverify file from minio given the bucket name and the file name
// Returns the file as a dto.File
func (m *MinioStorageHelper) Unverify(bucket string, filePath string, ctx context.Context) (*dto.File, error) {
	return m.setTags(bucket, filePath, "verified", "false", ctx)
}

// Check if file already exists in minio given the bucket name and the file name
// Returns a boolean
func (m *MinioStorageHelper) Exists(bucket string, filePath string, ctx context.Context) (bool, error) {
	_, err := m.client.StatObject(ctx, bucket, filePath, minio.StatObjectOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// Download file from minio given the bucket name and the file name
// Returns the reader of the file
func (m *MinioStorageHelper) Download(bucket string, filePath string, ctx context.Context) (*minio.Object, *minio.ObjectInfo, error) {
	if exists, _ := m.Exists(bucket, filePath, ctx); !exists {
		return nil, nil, NewFileDoesNotExistError(bucket, filePath) // TODO: add error code
	}
	obj, err := m.client.GetObject(ctx, bucket, filePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, err
	}
	f, err := obj.Stat()
	if err != nil {
		return nil, nil, err
	}
	return obj, &f, nil
}
