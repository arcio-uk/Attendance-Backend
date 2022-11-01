package utils

import (
	"sync"
)

type ReaderWriterLock struct {
	WriterWait  sync.WaitGroup
	Lock        sync.Mutex
	ReaderCount int
}
