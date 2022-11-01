package utils

import (
	"arcio/attendance-system/config"
	"database/sql"
	"sync"
	"testing"
	"time"
)

var conf, _ = config.LoadConfig()

func TestInitDatabase(t *testing.T) {
	_, err := InitDatabasePool(conf)

	if err != nil {
		t.Log("Failed to init database pool")
		t.Fail()
	}
}

const TEST_WRITES = 10

func TestWriteLater(t *testing.T) {
	var lock sync.Mutex
	counter := 0

	pool, err := InitDatabasePool(conf)
	if err != nil {
		t.Log("Failed to init database pool")
		t.Fail()
	}

	t.Log("Waiting for the writer thread to start maybe")
	time.Sleep(time.Millisecond * 100)

	i := 0
	t.Log("Testing for deadlocks")
	for i < TEST_WRITES {
		WriteLater(pool, func(database *sql.DB) {
			lock.Lock()
			counter++
			lock.Unlock()
		})

		i++
	}

	t.Log("Waiting 100ms for the \"writes\" to happen")
	time.Sleep(time.Millisecond * 100)

	if counter != TEST_WRITES {
		t.Logf("Expected %d writes to happen but only %d did\n", TEST_WRITES, counter)
		t.Fail()
	}
}

const TEST_WRITES_2 = 5000

func TestPerformance(t *testing.T) {
	var lock sync.Mutex
	counter := 0

	pool, err := InitDatabasePool(conf)
	if err != nil {
		t.Log("Failed to init database pool")
		t.Fail()
	}

	t.Log("Waiting for the writer thread to start maybe")
	time.Sleep(time.Millisecond * 100)

	i := 0
	t.Log("Testing for deadlocks")
	for i < TEST_WRITES_2 {
		WriteLater(pool, func(database *sql.DB) {
			lock.Lock()
			counter++
			lock.Unlock()
		})

		i++
	}

	t.Log("Waiting 100ms for the \"writes\" to happen")
	time.Sleep(time.Millisecond * 100)

	if counter != TEST_WRITES_2 {
		t.Logf("Expected %d writes to happen but only %d did\n", TEST_WRITES, counter)
		t.Fail()
	}
}
