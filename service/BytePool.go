package service

import "sync"

const bufferSize = 1024 * 1024

var bytePool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, bufferSize)
		return b
	},
}
