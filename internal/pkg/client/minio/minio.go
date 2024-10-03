package minio

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// IMinio defines the interface for interacting with a MinIO client.
type IMinio interface {
	// Read retrieves the content of the specified object as a slice of strings,
	// splitting the content by newline characters.
	Read(ctx context.Context, objectName string) ([]string, error)

	// Write uploads a slice of strings as a file to the specified object name in a MinIO bucket.
	// The lines are joined into a single string with newline characters before being uploaded.
	Write(ctx context.Context, objectName string, lines []string) error
}

// MinioClient wraps the MinIO client
type MinioClient struct {
	client     *minio.Client
	bucketName string
}

// NewMinioClient initializes a new MinioClient
func NewMinioClient(endpoint, accessKey, secretKey, bucketName string) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // Set to true if using HTTPS
	})
	if err != nil {
		return nil, err
	}

	return &MinioClient{client: client, bucketName: bucketName}, nil
}

// Read returns the content of the file as a slice of strings split by newline
func (m *MinioClient) Read(ctx context.Context, objectName string) ([]string, error) {
	// Download the object
	object, err := m.client.GetObject(ctx, m.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	// Read the content into a buffer
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, object); err != nil {
		return nil, err
	}

	// Split the content by newline
	lines := bytes.Split(buf.Bytes(), []byte{'\n'})

	// Convert [][]byte to []string
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = string(line)
	}

	return result, nil
}

// Write uploads a slice of strings as a file to MinIO
func (m *MinioClient) Write(ctx context.Context, objectName string, lines []string) error {
	// Join the lines into a single string with newline characters
	content := strings.Join(lines, "\n")

	// Create a reader from the content
	reader := strings.NewReader(content)

	// Upload the content to MinIO
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, reader, int64(len(content)), minio.PutObjectOptions{})
	return err
}
