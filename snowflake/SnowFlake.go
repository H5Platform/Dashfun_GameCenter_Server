package snowflake

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

const (
	workerIdBits uint8 = 10
	numberBits   uint8 = 12
	workerMax          = -1 ^ (-1 << workerIdBits)
	numberMax          = -1 ^ (-1 << numberBits)
	timeShift          = numberBits + workerIdBits
	workerShift        = numberBits
	startTime          = 1717529025000
)

type Worker struct {
	mu        sync.Mutex
	timestamp int64
	workerId  int64
	number    int64
}

func Must(worker *Worker, err error) *Worker {
	if err != nil {
		panic(err)
	}
	return worker
}

func GetWorker(workerId int64) (*Worker, error) {
	if workerId < 0 || workerId > workerMax {
		return nil, errors.New("worker id out of range")
	}
	return &Worker{
		timestamp: 0,
		workerId:  workerId,
		number:    0,
	}, nil
}

func (w *Worker) NextId() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now().UnixNano() / 1e6
	if w.timestamp == now {
		w.number++
		if w.number > numberMax {
			for now <= w.timestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		w.number = 0
		w.timestamp = now
	}
	id := (now-startTime)<<timeShift | (w.workerId << workerShift) | w.number
	return id
}

func (w *Worker) NextStrId() string {
	id := w.NextId()
	return strconv.FormatInt(id, 16)
}
