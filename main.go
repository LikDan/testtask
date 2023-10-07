package main

import (
	"errors"
	"fmt"
	"time"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

// A LifeMeaningless represents a meaninglessness of our life
type LifeMeaningless struct {
	id             int
	creationTime   time.Time // время создания
	executableTime time.Time // время выполнения
	err            error
	taskResult     TaskResult
}

type TaskResult = int

const (
	SUCCESS TaskResult = iota
	WRONG
)

func main() {
	superChan := make(chan LifeMeaningless, 10)
	go createTasks(superChan)

	doneTasks := make(chan LifeMeaningless)
	undoneTasks := make(chan LifeMeaningless)

	go executeTasks(superChan, doneTasks, undoneTasks)

	successResult := map[int]LifeMeaningless{}
	errorResult := map[int]LifeMeaningless{}

	go func() {
		for r := range doneTasks {
			successResult[r.id] = r
		}
		close(doneTasks)
	}()

	go func() {
		for r := range undoneTasks {
			errorResult[r.id] = r
		}
		close(undoneTasks)
	}()

	time.Sleep(time.Second * 3)

	println("Errors:")
	for r, t := range errorResult {
		println(r, t.err.Error())
	}

	println("Done task ids:")
	for r := range successResult {
		println(r)
	}
}

func createTasks(c chan LifeMeaningless) {
	go func() {
		for {
			creationTime := time.Now()
			var err error
			// вот такое условие появления ошибочных тасков
			// upd: replaced Nanosecond to Second - because every time nanosecond has 3 zeros after num.
			if (creationTime.Second())%2 == 1 {
				err = errors.New("some error occurred") // вот такое условие появления ошибочных тасков
			}

			c <- LifeMeaningless{
				creationTime: creationTime,
				err:          err,
				id:           creationTime.Nanosecond(),
			} // передаем таск на выполнение
		}
	}()
}

func proceedTask(a LifeMeaningless) LifeMeaningless {
	if a.err == nil {
		a.taskResult = SUCCESS
	} else {
		a.taskResult = WRONG
	}

	a.executableTime = time.Now()
	time.Sleep(time.Millisecond * 150)

	return a
}

func sortTask(a LifeMeaningless, doneTasks chan LifeMeaningless, undoneTasks chan LifeMeaningless) {
	if a.taskResult == SUCCESS {
		doneTasks <- a
	} else {
		a.err = fmt.Errorf("task id %d time %s, error %s", a.id, a.creationTime, a.err)
		undoneTasks <- a
	}
}

func executeTasks(tasksChan chan LifeMeaningless, doneTasks chan LifeMeaningless, undoneTasks chan LifeMeaningless) {
	func() {
		// получение тасков
		for t := range tasksChan {
			t = proceedTask(t)
			go sortTask(t, doneTasks, undoneTasks)
		}
		close(tasksChan)
	}()
}
