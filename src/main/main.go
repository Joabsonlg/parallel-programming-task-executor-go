package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
)

type TaskType int

const (
	READING TaskType = iota
	WRITING
)

type Task struct {
	ID       int
	Custo    float64
	TaskType TaskType
	Value    int
}

type Result struct {
	ID     int
	Result string
	Time   int64
}

type Executor struct {
	TaskQueue chan Task
	Results   chan Result
}

func NewExecutor(taskQueueSize, resultQueueSize int) *Executor {
	return &Executor{
		TaskQueue: make(chan Task, taskQueueSize),
		Results:   make(chan Result, resultQueueSize),
	}
}

func worker(tasks chan Task, sharedFile *os.File, results chan Result, rwMutex *sync.RWMutex, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		startTime := time.Now()
		if task.TaskType == WRITING {
			write(task, sharedFile, rwMutex)
		} else {
			read(task, sharedFile, rwMutex)
		}
		endTime := time.Now()
		executionTime := endTime.Sub(startTime).Nanoseconds()
		results <- Result{ID: task.ID, Time: executionTime}
	}
}

func write(task Task, sharedFile *os.File, rwMutex *sync.RWMutex) {
	time.Sleep(time.Duration(task.Custo*1000) * time.Millisecond)

	rwMutex.Lock()
	defer rwMutex.Unlock()

	_, err := sharedFile.Seek(0, 0)
	if err != nil {
		log.Println(err)
		return
	}

	var existingValue int
	_, err = fmt.Fscanln(sharedFile, &existingValue)
	if err != nil {
		log.Println(err)
		return
	}

	newValue := existingValue + task.Value
	sharedFile.Truncate(0)
	sharedFile.Seek(0, 0)
	_, err = sharedFile.WriteString(fmt.Sprintf("%d\n", newValue))
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("W - Task %d : %d\n", task.ID, newValue)
}

func read(task Task, sharedFile *os.File, rwMutex *sync.RWMutex) {
	time.Sleep(time.Duration(task.Custo*1000) * time.Millisecond)

	rwMutex.RLock()
	defer rwMutex.RUnlock()

	_, err := sharedFile.Seek(0, 0)
	if err != nil {
		log.Println(err)
		return
	}

	var value string
	_, err = fmt.Fscanln(sharedFile, &value)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("R - Task %d : %s\n", task.ID, value)
}

func main() {
	var nTasksAmount, tWorkersThreads, eWritingTasks int
	fmt.Print("Digite o valor de N: ")
	fmt.Scan(&nTasksAmount)

	fmt.Print("Digite o valor de T: ")
	fmt.Scan(&tWorkersThreads)

	fmt.Print("Digite o valor de E: ")
	fmt.Scan(&eWritingTasks)

	tasksAmount := int(math.Pow(10, float64(nTasksAmount)))
	writingTasks := int(float64(eWritingTasks) / 100 * float64(tasksAmount))
	readingTasks := tasksAmount - writingTasks

	executor := NewExecutor(tasksAmount, tasksAmount)

	// Gerar tarefas de forma misturada
	taskId := 0
	tasks := make([]Task, 0, tasksAmount)

	for i := 0; i < writingTasks; i++ {
		tasks = append(tasks, Task{ID: taskId, Custo: rand.Float64() * 0.01, TaskType: WRITING, Value: rand.Intn(11)})
		taskId++
	}

	for i := 0; i < readingTasks; i++ {
		tasks = append(tasks, Task{ID: taskId, Custo: rand.Float64() * 0.01, TaskType: READING, Value: rand.Intn(11)})
		taskId++
	}

	// Embaralhar as tarefas
	rand.Shuffle(len(tasks), func(i, j int) {
		tasks[i], tasks[j] = tasks[j], tasks[i]
	})

	// Enfileirar as tarefas embaralhadas
	for _, task := range tasks {
		executor.TaskQueue <- task
	}
	close(executor.TaskQueue) // Fechando TaskQueue após inserção de todas as tarefas

	// Criar arquivo compartilhado
	sharedFile, err := os.Create("shared_file.txt")
	if err != nil {
		log.Fatalf("Erro ao criar arquivo compartilhado: %v", err)
	}
	defer sharedFile.Close()

	// Escrever valor inicial no arquivo
	_, err = sharedFile.WriteString("0\n")
	if err != nil {
		log.Fatalf("Erro ao escrever no arquivo compartilhado: %v", err)
	}
	sharedFile.Sync()

	var wg sync.WaitGroup
	rwMutex := &sync.RWMutex{}

	for i := 0; i < tWorkersThreads; i++ {
		wg.Add(1)
		go worker(executor.TaskQueue, sharedFile, executor.Results, rwMutex, &wg)
	}

	go func() {
		wg.Wait()
		close(executor.Results) // Fechando Results após todas as goroutines terminarem
	}()

	// Coletar resultados
	for range executor.Results {
		// Sem print aqui, já que os detalhes são impressos nas funções read e write
	}

	fmt.Println("Fim da execução.")
}
