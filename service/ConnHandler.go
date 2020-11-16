package service

import (
	"fmt"
	"net"
	"time"
)

const (
	ClientTimeout = time.Second * 15
)

func HandleConn(conn net.Conn) {
	parser := NewParser(conn)
	defer parser.Release()
	parser.Run()

	reqQueue := NewRequestQueue()
	defer reqQueue.Release()
	reqQueue.Run()

loop:
	for {
		select {
		case <-parser.ErrConn:
			break loop
		case <-parser.ErrCh:
			reqQueue.Queue <- &BadRequest{
				Id:   parser.Id,
				Conn: parser.Conn,
			}
			break loop
		case <-time.After(ClientTimeout):
			if len(parser.Data) > 0 {
				reqQueue.Queue <- &RequestTimeout{
					Id:   parser.Id,
					Conn: parser.Conn,
				}
			}
			if reqQueue.Done() {
				break loop
			}
		case req := <-parser.Done:
			reqQueue.Queue <- req
			fmt.Println("done one request:", time.Now())
		}
	}
}
