package main

// Executor Estrutura para gerenciar filas de tarefas e resultados
type Executor struct {
	TaskQueue chan int32 // Mudança para canal de índices de tarefas
	Results   chan Result
	TaskData  *TaskDataStore // Referência ao TaskDataStore
}

// NewExecutor Função para criar um novo Executor com tamanhos de fila especificados
func NewExecutor(taskQueueSize, resultQueueSize int, taskData *TaskDataStore) *Executor {
	return &Executor{
		TaskQueue: make(chan int32, taskQueueSize),
		Results:   make(chan Result, resultQueueSize),
		TaskData:  taskData,
	}
}
