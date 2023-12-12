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

func main() {
	var nTasksAmount, tWorkersThreads, eWritingTasks int
	fmt.Print("Digite o valor de N: ")
	fmt.Scan(&nTasksAmount)

	fmt.Print("Digite o valor de T: ")
	fmt.Scan(&tWorkersThreads)

	fmt.Print("Digite o valor de E: ")
	fmt.Scan(&eWritingTasks)

	//nValues := []int{9}
	//tValues := []int{1, 16, 256}
	//eValues := []int{0, 40}
	//
	//for _, nTasksAmount := range nValues {
	//	for _, tWorkersThreads := range tValues {
	//		for _, eWritingTasks := range eValues {
	tasksAmount := int(math.Pow(10, float64(nTasksAmount)))
	writingTasks := int(float64(eWritingTasks) / 100 * float64(tasksAmount))

	fmt.Printf("N: %d, T: %d, E: %d\n", nTasksAmount, tWorkersThreads, eWritingTasks)

	// Inicializando TaskDataStore
	taskDataStore := &TaskDataStore{
		ids:       make([]int32, tasksAmount),
		custos:    make([]int16, tasksAmount),
		taskTypes: make([]bool, tasksAmount),
		values:    make([]byte, tasksAmount),
	}

	// Gerar tarefas
	for i := 0; i < tasksAmount; i++ {
		taskDataStore.ids[i] = int32(i)
		taskDataStore.SetCusto(i, rand.Float32()*0.01)
		taskDataStore.taskTypes[i] = determineTaskType(i, writingTasks)
		taskDataStore.values[i] = byte(rand.Intn(11)) // ajuste conforme necessário
	}

	// Embaralhar as tarefas
	//rand.Shuffle(len(taskDataStore.ids), func(i, j int) {
	//	taskDataStore.ids[i], taskDataStore.ids[j] = taskDataStore.ids[j], taskDataStore.ids[i]
	//	taskDataStore.custos[i], taskDataStore.custos[j] = taskDataStore.custos[j], taskDataStore.custos[i]
	//	taskDataStore.taskTypes[i], taskDataStore.taskTypes[j] = taskDataStore.taskTypes[j], taskDataStore.taskTypes[i]
	//	taskDataStore.values[i], taskDataStore.values[j] = taskDataStore.values[j], taskDataStore.values[i]
	//})

	// Inicializar Executor
	executor := NewExecutor(tasksAmount, tasksAmount, taskDataStore)

	// Enfileirar os índices das tarefas embaralhadas
	for i := 0; i < tasksAmount; i++ {
		executor.TaskQueue <- int32(i)
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

	// Iniciar processamento
	fmt.Println("Iniciando processamento...")
	startTime := time.Now()

	var wg sync.WaitGroup
	rwMutex := &sync.RWMutex{}

	for i := 0; i < tWorkersThreads; i++ {
		wg.Add(1)
		go worker(i, executor, sharedFile, rwMutex, &wg)
	}

	go func() {
		wg.Wait()
		close(executor.Results) // Fechando Results após todas as goroutines terminarem

		// Registrar o fim do processamento
		endTime := time.Now()
		totalTime := endTime.Sub(startTime)
		fmt.Printf("Tempo total de processamento (nanossegundos): %d\n", totalTime.Nanoseconds())

		// Create and open the file
		fileName := fmt.Sprintf("N%dT%dE%d.txt", nTasksAmount, tWorkersThreads, eWritingTasks)
		fmt.Println("Criando arquivo:", fileName)
		resultFile, err := os.Create(fileName)
		if err != nil {
			log.Fatalf("Failed to create file: %v", err)
		}
		defer resultFile.Close()

		// Write the processing time into the file
		_, err = resultFile.WriteString(fmt.Sprintf("Total processing time (nanoseconds): %d\n", totalTime.Nanoseconds()))
		if err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}

		// Check if the write was successful
		err = resultFile.Sync()
		if err != nil {
			log.Fatalf("Failed to sync file: %v", err)
		}
	}()

	for range executor.Results {
		//fmt.Println("Tarefa processada:", <-executor.Results)
	}

	fmt.Println("Fim da execução.\n")
	//		}
	//	}
	//}
}
