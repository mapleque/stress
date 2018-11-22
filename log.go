package main

import (
	"encoding/json"
	"fmt"
	"time"
)

func (this *Stress) log() {
	fmt.Printf("stress start at %s\n", this.startAt.Format(time.RFC3339))
	b, _ := json.MarshalIndent(this, "", "\t")
	fmt.Printf("running with following configure:\n%s\n", string(b))
	fmt.Println("Here is the realtime result:")
	fmt.Println("================================================================")
	fmt.Println("run\tthread\tqps\ttotal\tdoing\tsucc\tfail\tlant\tsucc_r\tfail_r")
	go func() {
		timer := time.NewTicker(time.Duration(this.LogInterval) * time.Second)
		for c := range timer.C {
			if this.isPause {
				return
			}
			this.logRecord(c)
		}
	}()
}

func (this *Stress) logover() {
	fmt.Println("----------------------------------------------------------------")
	this.logRecord(this.stopAt)
	fmt.Println("run\tthread\tqps\ttotal\tdoing\tsucc\tfail\tlant\tsucc_r\tfail_r")
	fmt.Println("================================================================")
	fmt.Println("Stress testing finished!")
	fmt.Printf("Find statistic info is in %sinfo.log\n", this.LogPath)
	fmt.Printf("Find error message is in %serror.log\n", this.LogPath)
	this.logInfoHandle.Close()
	this.logErrorHandle.Close()
}

func (this *Stress) logInfo(message interface{}) {
	fmt.Fprintln(this.logInfoHandle, message)
}

func (this *Stress) logError(message interface{}) {
	fmt.Fprintln(this.logErrorHandle, message)
}

func (this *Stress) logRecord(c time.Time) {
	// read before for decrease racing
	total, succ, fail, latency := this.snapshot()
	doing := total - succ - fail

	var succRate, failRate float64
	var avg int64
	if total > 0 {
		succRate = float64(succ) / float64(total-doing)
		failRate = float64(fail) / float64(total-doing)
		// avg response time
		avg = latency / int64(total-doing) / int64(time.Millisecond)
	}
	var qps, run int64
	// total running time
	run = (c.Sub(this.startAt).Nanoseconds() - this.totalPauseTime) / int64(time.Second)
	qps = (total - doing) / run
	message := fmt.Sprintf(
		"%ds\t%d\t%d\t%d\t%d\t%d\t%d\t%dms\t%.2f%%\t%.2f%%",
		run,
		this.currentThreadNumber,
		qps,
		total,
		doing,
		succ,
		fail,
		avg,
		succRate*100,
		failRate*100,
	)
	fmt.Println(message)
	this.logInfo(message)
	if failRate > this.MaxFailedRatio {
		this.Stop()
	}
}

func (this *Stress) snapshot() (total, succ, fail, latency int64) {
	this.snapshotLock.RLock()
	defer this.snapshotLock.RUnlock()
	total, succ, fail, latency = this.totalRequestNumber,
		this.successRequestNumber,
		this.failedRequestNumber,
		this.totalResponseTime
	return
}

func (this *Stress) addTotal() {
	this.snapshotLock.Lock()
	defer this.snapshotLock.Unlock()
	this.totalRequestNumber += 1
}

func (this *Stress) addSuccess() {
	this.snapshotLock.Lock()
	defer this.snapshotLock.Unlock()
	this.successRequestNumber += 1
}

func (this *Stress) addFailed() {
	this.snapshotLock.Lock()
	defer this.snapshotLock.Unlock()
	this.failedRequestNumber += 1
}

func (this *Stress) addLatency(latency int64) {
	this.snapshotLock.Lock()
	defer this.snapshotLock.Unlock()
	this.totalResponseTime += latency
}
