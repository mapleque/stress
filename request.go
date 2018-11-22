package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// request request target and check response
func (this *Stress) request() {
	client := &http.Client{
		Timeout: time.Duration(this.Timeout) * time.Millisecond,
	}
	var body io.Reader
	if len(this.Body) > 0 {
		body = bytes.NewReader([]byte(this.Body))
	}
	req, err := http.NewRequest(
		this.Method,
		this.Url,
		body,
	)
	for _, h := range this.Header {
		req.Header.Add(h.key, h.value)
	}

	// request will send
	start := time.Now()
	this.addTotal()
	resp, err := client.Do(req)

	latency := time.Now().Sub(start).Nanoseconds()
	this.addLatency(latency)

	if err != nil {
		this.addFailed()
		this.logError(err.Error())
		return
	}
	defer resp.Body.Close()

	if this.Response != "" {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			this.addFailed()
			this.logError(err.Error())
			return
		}
		if string(body) != this.Response {
			this.addFailed()
			this.logError(string(body))
			return
		}
		if resp.StatusCode != this.Status {
			this.addFailed()
			this.logError(resp.StatusCode)
			return
		}
	}
	this.addSuccess()
}
