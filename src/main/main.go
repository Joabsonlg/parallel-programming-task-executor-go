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

// TaskType Declaração de um tipo personalizado para diferenciar tarefas de leitura e escrita
type TaskType int

const (
	READING TaskType = iota
	WRITING
)

// Task Estrutura para armazenar detalhes de uma tarefa
type Task struct {
	ID       int
	Custo    float64
	TaskType TaskType
	Value    int
}

// Result Estrutura para armazenar o resultado de uma tarefa processada
type Result struct {
	ID     int
	Result string
	Time   int64
}

// Executor Estrutura para gerenciar filas de tarefas e resultados
type Executor struct {
	TaskQueue chan Task
	Results   chan Result
}

// NewExecutor Função para criar um novo Executor com tamanhos de fila especificados
func NewExecutor(taskQueueSize, resultQueueSize int) *Executor {
	return &Executor{
		TaskQueue: make(chan Task, taskQueueSize),
		Results:   make(chan Result, resultQueueSize),
	}
}

// Função worker para processar tarefas. Cada worker executa em uma goroutine separada.
func worker(workerID int, tasks chan Task, sharedFile *os.File, results chan Result, rwMutex *sync.RWMutex, wg *sync.WaitGroup) {
	defer wg.Done() // Sinaliza a conclusão da goroutine ao WaitGroup ao retornar
	for task := range tasks {
		startTime := time.Now()
		//fmt.Printf("Worker %d está processando a tarefa %d\n", workerID, task.ID)
		if task.TaskType == WRITING {
			write(task, sharedFile, rwMutex)
		} else {
			read(task, sharedFile, rwMutex)
		}
		endTime := time.Now()
		executionTime := endTime.Sub(startTime).Nanoseconds()
		results <- Result{ID: task.ID, Time: executionTime} // Envia o resultado para o canal Results
	}
}

// Função write para processar tarefas de escrita
func write(task Task, sharedFile *os.File, rwMutex *sync.RWMutex) {
	time.Sleep(time.Duration(task.Custo*1000) * time.Millisecond)

	rwMutex.Lock()         // Bloqueia para escrita
	defer rwMutex.Unlock() // Desbloqueia após a conclusão da função

	_, err := sharedFile.Seek(0, 0) // Posiciona o ponteiro do arquivo no início
	if err != nil {
		log.Println(err)
		return
	}

	var existingValue int
	_, err = fmt.Fscanln(sharedFile, &existingValue) // Lê o valor existente no arquivo
	if err != nil {
		log.Println(err)
		return
	}

	newValue := existingValue + task.Value
	sharedFile.Truncate(0)                                         // Limpa o conteúdo do arquivo
	sharedFile.Seek(0, 0)                                          // Reposiciona o ponteiro no início
	_, err = sharedFile.WriteString(fmt.Sprintf("%d\n", newValue)) // Escreve o novo valor
	if err != nil {
		log.Println(err)
		return
	}

	//fmt.Printf("W - Task %d : %d\n", task.ID, newValue)
}

// Função read para processar tarefas de leitura
func read(task Task, sharedFile *os.File, rwMutex *sync.RWMutex) {
	time.Sleep(time.Duration(task.Custo*1000) * time.Millisecond)

	rwMutex.RLock()         // Bloqueia para leitura
	defer rwMutex.RUnlock() // Desbloqueia após a conclusão da função

	_, err := sharedFile.Seek(0, 0) // Posiciona o ponteiro do arquivo no início
	if err != nil {
		log.Println(err)
		return
	}

	var value string
	_, err = fmt.Fscanln(sharedFile, &value) // Lê o valor do arquivo
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("R - Task %d : %s\n", task.ID, value)
}

// Função main: ponto de entrada do programa
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
	tasks := make([]Task, 0, tasksAmount) // Cria um slice de tarefas vazio com tamanho inicial 0 e capacidade tasksAmount

	for i := 0; i < writingTasks; i++ { // Adiciona tarefas de escrita
		tasks = append(tasks, Task{ID: taskId, Custo: rand.Float64() * 0.01, TaskType: WRITING, Value: rand.Intn(11)})
		taskId++
	}

	for i := 0; i < readingTasks; i++ { // Adiciona tarefas de leitura
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

	startTime := time.Now()

	var wg sync.WaitGroup
	rwMutex := &sync.RWMutex{}

	for i := 0; i < tWorkersThreads; i++ {
		wg.Add(1)
		go worker(i, executor.TaskQueue, sharedFile, executor.Results, rwMutex, &wg)
	}

	go func() {
		wg.Wait()
		close(executor.Results) // Fechando Results após todas as goroutines terminarem

		// Registrar o fim do processamento
		endTime := time.Now()

		// Calcular e exibir o tempo total
		totalTime := endTime.Sub(startTime)
		fmt.Printf("Tempo total de processamento: %s\n", totalTime)
	}()

	// Coletar resultados
	for range executor.Results {
		// Sem print aqui, já que os detalhes são impressos nas funções read e write
	}

	fmt.Println("Fim da execução.")
}
