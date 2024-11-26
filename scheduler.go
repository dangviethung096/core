package core

import (
	"time"
)

var done = make(chan bool)
var w *worker //you can declare multi worker here

type task struct {
	ID        uint64 `db:"id"`
	QueueName string `db:"queue_name"`
	Data      []byte `db:"data"`
	Done      bool   `db:"done"`
	LoopIndex uint64 `db:"loop_index"`
	LoopCount uint64 `db:"loop_count"`
	Next      int64  `db:"next"`
	Interval  int64  `db:"interval"`
	Source    string `db:"source"`
	StartTime string `db:"start_time"`
}

func startScheduler(interval time.Duration) {
	w = NewWorker()
	w.Start(0, interval)
}

func stopScheduler() {
	done <- true
}

func GetBucket(time time.Time) int64 {
	return time.Unix() / int64(Config.Scheduler.BucketSize)
}

func StopSchedulerForTest() {
	stopScheduler()
}
