package main

type TaskDataStore struct {
	ids       []int32
	custos    []int16
	taskTypes []bool
	values    []byte
}

func determineTaskType(index, writingTasks int) bool {
	return index < writingTasks
}

func (t *TaskDataStore) SetCusto(index int, custo float32) {
	if custo < 0 || custo > 0.01 {
		return
	}
	t.custos[index] = int16(custo * 10000)
}

func (t *TaskDataStore) GetCusto(index int32) float32 {
	return float32(t.custos[index]) / 10000
}
