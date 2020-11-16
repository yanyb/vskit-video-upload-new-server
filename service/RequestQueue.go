package service

import (
	"errors"
	"github.com/vskit-tv/vlog/log"
	"sync"
	"time"
)

type RequestQueue struct {
	sync.Mutex
	Queue chan Request
	Quit  chan struct{}
	total int
}

func (rq *RequestQueue) Done() bool {
	rq.Lock()
	defer rq.Unlock()
	return rq.total == 0
}

func NewRequestQueue() *RequestQueue {
	return &RequestQueue{
		Queue: make(chan Request, 32),
	}
}

func (rq *RequestQueue) Release() {
	rq.Quit <- struct{}{}
}

func (rq *RequestQueue) increase() {
	rq.Lock()
	defer rq.Unlock()
	rq.total++
}

func (rq *RequestQueue) decrease() {
	rq.Lock()
	defer rq.Unlock()
	rq.total--
}

func (rq *RequestQueue) Run() {
	go func() {
		for {
			select {
			case req := <-rq.Queue:
				rq.increase()
				resp := req.Do()
				if err := sendResponse(req, resp); err != nil {
					log.Errorf("send response err = %+v", err)
				}
				rq.decrease()
			case <-rq.Quit:
				return
			}
		}
	}()
}

func sendResponse(req Request, resp *Response) error {
	send := 0
	content := []byte(resp.String())
	total := len(content)
	// limit 5s
	startTime := time.Now().Add(time.Second * 5)
	for {
		n, err := req.GetConn().Write(content[send:])
		if err != nil {
			return err
		}
		send += n
		if send == total {
			break
		}
		if time.Now().After(startTime) {
			return errors.New("send time out")
		}
	}
	return nil
}
