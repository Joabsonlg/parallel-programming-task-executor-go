package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewExecutor(t *testing.T) {
	taskDataStore := &TaskDataStore{
		custos: make([]int16, 1),
	}
	executor := NewExecutor(1, 1, taskDataStore)
	assert.NotNil(t, executor, "NewExecutor should return a non-nil executor")
}

func TestExecutorTaskQueue(t *testing.T) {
	taskDataStore := &TaskDataStore{
		custos: make([]int16, 1),
	}
	executor := NewExecutor(1, 1, taskDataStore)
	executor.TaskQueue <- 1
	taskIndex := <-executor.TaskQueue
	assert.Equal(t, int32(1), taskIndex, "TaskQueue should return the same task index that was inserted")
}

func TestExecutorResults(t *testing.T) {
	taskDataStore := &TaskDataStore{
		custos: make([]int16, 1),
	}
	executor := NewExecutor(1, 1, taskDataStore)
	executor.Results <- Result{ID: 1, Time: 1, Value: "1"}
	result := <-executor.Results
	assert.Equal(t, Result{ID: 1, Time: 1, Value: "1"}, result, "Results should return the same result that was inserted")
}

func TestTaskDataStoreSetAndGetCusto(t *testing.T) {
	taskDataStore := &TaskDataStore{
		custos: make([]int16, 1),
	}
	taskDataStore.SetCusto(0, 0.005)
	assert.Equal(t, float32(0.005), taskDataStore.GetCusto(0), "GetCusto should return the same cost that was set by SetCusto")
}

func TestTaskDataStoreSetCustoInvalid(t *testing.T) {
	taskDataStore := &TaskDataStore{
		custos: make([]int16, 1),
	}
	taskDataStore.SetCusto(0, 0.02)
	assert.Equal(t, float32(0), taskDataStore.GetCusto(0), "GetCusto should return 0 when an invalid cost was set")
}

func TestTaskDataStoreSetCustoNegative(t *testing.T) {
	taskDataStore := &TaskDataStore{
		custos: make([]int16, 1),
	}
	taskDataStore.SetCusto(0, -0.005)
	assert.Equal(t, float32(0), taskDataStore.GetCusto(0), "GetCusto should return 0 when a negative cost was set")
}

func TestDetermineTaskTypeWriting(t *testing.T) {
	assert.Equal(t, true, determineTaskType(0, 1), "determineTaskType should return true when index is less than writingTasks")
}

func TestDetermineTaskTypeReading(t *testing.T) {
	assert.Equal(t, false, determineTaskType(1, 1), "determineTaskType should return false when index is not less than writingTasks")
}
