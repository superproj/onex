package fake

import (
	"context"
	"strings"
	"sync"
)

const (
	FakeObjectName    = "llm/test.json"
	FakeObjectContent = `The quick brown fox jumps over the lazy dog.\nArtificial intelligence is transforming the world.\nNatural language processing enables machines to understand human language.\nDeep learning models require large amounts of data for training.\nData preprocessing is a crucial step in machine learning.\nSupervised learning involves training a model on labeled data.\nUnsupervised learning allows models to find patterns in unlabeled data.\nReinforcement learning is used in scenarios where an agent learns by interacting with an environment.\nThe future of AI holds great potential for various industries.\nEthics in AI is an important consideration for developers and researchers.`
)

// FakeMinioClient is a mock implementation of MinioClient for testing
type FakeMinioClient struct {
	data       map[string]string // Stores data keyed by objectName
	bucketName string
	mu         sync.RWMutex // Mutex for synchronizing access to data
}

// NewFakeMinioClient initializes a new FakeMinioClient
func NewFakeMinioClient(bucketName string) (*FakeMinioClient, error) {
	fake := &FakeMinioClient{
		data:       make(map[string]string),
		bucketName: bucketName,
	}
	fake.InitializeData(FakeObjectName, FakeObjectContent)
	return fake, nil
}

// Read returns the content of the fake file as a slice of strings split by newline
func (f *FakeMinioClient) Read(ctx context.Context, objectName string) ([]string, error) {
	f.mu.RLock()         // Acquire a read lock
	defer f.mu.RUnlock() // Ensure the lock is released

	content := FakeObjectContent
	if value, exists := f.data[objectName]; exists {
		content = value
	}

	// Split the content by newline
	lines := strings.Split(content, "\n")
	return lines, nil
}

// Write uploads a slice of strings as a file to the fake MinIO
func (f *FakeMinioClient) Write(ctx context.Context, objectName string, lines []string) error {
	f.mu.Lock()         // Acquire a write lock
	defer f.mu.Unlock() // Ensure the lock is released

	// Join the lines into a single string with newline characters
	content := strings.Join(lines, "\n")

	// Store the content in the fake data map
	f.data[objectName] = content
	return nil
}

// InitializeData populates the fake MinIO client with initial object data
func (f *FakeMinioClient) InitializeData(objectName string, content string) {
	f.mu.Lock()         // Acquire a write lock
	defer f.mu.Unlock() // Ensure the lock is released

	f.data[objectName] = content
}
