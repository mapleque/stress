package main

import (
	"os"
	"sync"
	"time"
)

const (
	MODE_DYNAMIC = "dynamic"
	MODE_STATIC  = "static"
)

type Stress struct {
	// request settings
	Url     string  `json:"url"`
	Method  string  `json:"method"`
	Body    string  `json:"body"`
	Header  fheader `json:"header"`
	Timeout int     `json:"timeout"` // millisecond

	// response settings
	Status         int     `json:"status"`
	Response       string  `json:"response"`
	MaxFailedRatio float64 `json:"max-failed-ratio"`

	// runtime settings
	Mode     string `json:"mode"` // MODE_*
	Step     int    `json:"step"`
	Stay     int    `json:"stay"` // second
	Thread   int    `json:"thread"`
	Interval int    `json:"interval"` // millisecond

	// log settings
	LogInterval int    `json:"log-interval"`
	LogPath     string `json:"log-path"`

	// config file
	configFilePath string `json:"-"`

	// runtime var
	isRunning      bool
	isPause        bool
	done           chan bool
	exit           chan bool
	pauseCond      *sync.Cond
	stopCond       *sync.Cond
	snapshotLock   *sync.RWMutex
	logInfoHandle  *os.File
	logErrorHandle *os.File

	// result
	totalRequestNumber   int64
	successRequestNumber int64
	failedRequestNumber  int64
	currentThreadNumber  int64
	totalResponseTime    int64 // nanosecond

	totalPauseTime int64 // nanosecond
	startAt        time.Time
	stopAt         time.Time
	lastPauseAt    time.Time
}

func New() *Stress {
	return &Stress{
		pauseCond:    sync.NewCond(new(sync.Mutex)),
		stopCond:     sync.NewCond(new(sync.Mutex)),
		snapshotLock: new(sync.RWMutex),
		done:         make(chan bool, 1),
		exit:         make(chan bool, 1),
	}
}

// Start start to request target
func (this *Stress) Start() {
	if this.Thread < 1 {
		panic("thread must > 0")
	}
	if this.isRunning {
		return
	}
	this.isRunning = true
	go this.run()
}

func (this *Stress) run() {
	this.startAt = time.Now()
	this.log()

	var wg sync.WaitGroup

	wg.Add(this.Thread)
	this.currentThreadNumber = int64(this.Thread)
	for i := 0; i < this.Thread; i++ {
		go func() {
			defer wg.Done()
			this.working()
		}()
	}

	switch this.Mode {
	case MODE_STATIC:
		// do nothing
	case MODE_DYNAMIC:
		// increase threads by step
		if this.Stay < 1 {
			panic("stay must > 0")
		}
		if this.Step < 1 {
			panic("step must > 0")
		}
		go func() {
			timer := time.NewTicker(time.Duration(this.Stay) * time.Second)
			for range timer.C {
				if this.isRunning && !this.isPause {
					wg.Add(this.Step)
					this.currentThreadNumber += int64(this.Step)
					for i := 0; i < this.Step; i++ {
						go func() {
							defer wg.Done()
							this.working()
						}()
					}
				}
			}
		}()
	}
	wg.Wait()
	this.stopAt = time.Now()
	this.done <- true
}

func (this *Stress) waitingDone() {
	<-this.done
	this.logover()
	this.exit <- true
}

// Stop stop all
func (this *Stress) Stop() {
	if !this.isRunning {
		return
	}
	if !this.isPause {
		this.Pause()
	}
	this.isRunning = false
	this.stopCond.Broadcast()
	this.waitingDone()
}

// Pause pause and waiting for recover
func (this *Stress) Pause() {
	if this.isPause {
		return
	}
	this.isPause = true
	this.lastPauseAt = time.Now()
}

// Recover continue on pause point
func (this *Stress) Recover() {
	if !this.isPause {
		return
	}
	this.isPause = false
	this.pauseCond.Broadcast()
	this.totalPauseTime += time.Now().Sub(this.lastPauseAt).Nanoseconds()
}

func (this *Stress) working() {
	go func() {
		for {
			if this.isPause {
				this.pauseCond.L.Lock()
				this.pauseCond.Wait()
				this.pauseCond.L.Unlock()
			}
			this.request()
			if this.Interval > 0 {
				time.Sleep(time.Duration(this.Interval) * time.Millisecond)
			}
		}
	}()
	this.stopCond.L.Lock()
	defer this.stopCond.L.Unlock()
	this.stopCond.Wait()
}
