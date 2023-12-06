package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Função worker para processar tarefas. Cada worker executa em uma goroutine separada.
func worker(workerID int, executor *Executor, sharedFile *os.File, rwMutex *sync.RWMutex, wg *sync.WaitGroup) {
	defer wg.Done()
	for taskIndex := range executor.TaskQueue {
		startTime := time.Now()
		taskID := executor.TaskData.ids[taskIndex]
		taskValue := ""

		if executor.TaskData.taskTypes[taskIndex] {
			//write(taskIndex, executor.TaskData, sharedFile, rwMutex)
		} else {
			s, err := read(taskIndex, executor.TaskData, sharedFile, rwMutex)
			if err != nil {
				log.Fatal(err)
			}
			taskValue = s
		}

		endTime := time.Now()
		executionTime := endTime.Sub(startTime).Nanoseconds()
		executor.Results <- Result{ID: taskID, Time: int32(executionTime), Value: taskValue}
	}
}

// Função write para processar tarefas de escrita
func write(taskIndex int32, taskData *TaskDataStore, sharedFile *os.File, rwMutex *sync.RWMutex) {
	rwMutex.Lock()
	defer rwMutex.Unlock()

	time.Sleep(time.Duration(taskData.GetCusto(taskIndex)) * time.Millisecond)

	_, err := sharedFile.Seek(0, 0)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	var existingValue int
	_, err = fmt.Fscanln(sharedFile, &existingValue)
	if err != nil {
		log.Fatal(err)
	}

	newValue := existingValue + int(taskData.values[taskIndex])
	sharedFile.Truncate(0)
	sharedFile.Seek(0, 0)
	_, err = sharedFile.WriteString(fmt.Sprintf("%d\n", newValue))
	if err != nil {
		log.Fatal(err)
	}
}

// Função read para processar tarefas de leitura
func read(taskIndex int32, taskData *TaskDataStore, sharedFile *os.File, rwMutex *sync.RWMutex) (string, error) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	time.Sleep(time.Duration(taskData.GetCusto(taskIndex)) * time.Millisecond)

	_, err := sharedFile.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("erro ao posicionar o cursor no início do arquivo: %v", err)
	}

	scanner := bufio.NewScanner(sharedFile)
	if scanner.Scan() {
		return scanner.Text(), nil
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("erro ao ler a linha: %v", err)
	}

	return "", nil
}
