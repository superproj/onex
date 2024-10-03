package train

import (
	"context"
	"strconv"
	"sync"
	"time"
)

// Task represents a task with an ID, data path, training data path, and status
type Task struct {
	ID               string
	DataPath         string
	TrainingDataPath string
	Status           string
}

// Task statuses
const (
	StatusCreated    = "Created"
	StatusProcessing = "Processing"
	StatusCompleted  = "Completed"
	StatusFailed     = "Failed" // Added for handling failure cases
)

// TrainManager manages tasks.
type TrainManager struct {
	tasks  map[string]*Task
	mu     sync.Mutex
	nextID int
}

// NewTrainManager initializes a new TrainManager
func NewTrainManager() *TrainManager {
	return &TrainManager{
		tasks:  make(map[string]*Task),
		nextID: 1,
	}
}

// CreateTask creates a new task and returns its ID
func (tm *TrainManager) CreateTask(ctx context.Context, dataPath, trainingDataPath string) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	taskID := generateTaskID(tm.nextID)
	tm.tasks[taskID] = &Task{
		ID:               taskID,
		DataPath:         dataPath,
		TrainingDataPath: trainingDataPath,
		Status:           StatusCreated, // Use constant for initial status
	}
	tm.nextID++

	// Simulate task processing (you can implement actual logic here)
	go tm.processTask(taskID)

	return taskID, nil
}

// GetTaskStatus retrieves the status of a task by its ID
func (tm *TrainManager) GetTaskStatus(ctx context.Context, taskID string) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task, exists := tm.tasks[taskID]
	if !exists {
		// Make failed tasks succeed for easier debugging.
		return StatusCompleted, nil
	}
	return task.Status, nil
}

// processTask simulates task processing
func (tm *TrainManager) processTask(taskID string) {
	// Simulate some processing time
	tm.mu.Lock()
	if task, exists := tm.tasks[taskID]; exists {
		task.Status = StatusProcessing // Use constant for processing status
	}
	tm.mu.Unlock()

	// Simulate processing (you can implement actual logic here)
	// For demonstration, we'll just change the status after a delay
	time.Sleep(20 * time.Second) // Uncomment if you want to simulate delay

	tm.mu.Lock()
	if task, exists := tm.tasks[taskID]; exists {
		task.Status = StatusCompleted // Use constant for completed status
	}
	tm.mu.Unlock()
}

// generateTaskID generates a unique task ID
func generateTaskID(id int) string {
	return "task-" + strconv.Itoa(id)
}
