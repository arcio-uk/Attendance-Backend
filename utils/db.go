package utils

import (
	"arcio/attendance-system/config"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"sync"
	"time"
)

const WRITER_POLL_TIME = 10 * time.Millisecond

type DatabaseRunnable func(*sql.DB)
type DatabasePool struct {
	Database      *sql.DB
	PendingWrites []DatabaseRunnable
	WriteLock     sync.Mutex // Used to lock the PendingWrites array
}

func WriterThread(pool *DatabasePool) {
	log.Println("Started writer thread.")

	for true {
		pool.WriteLock.Lock() // Lock the pending writes array then copy the contents to a cache then unlock
		if len(pool.PendingWrites) > 0 {

			// Move the array then, change it to an empty one
			LocalWrites := pool.PendingWrites
			pool.PendingWrites = make([]DatabaseRunnable, 0)
			pool.WriteLock.Unlock() // Unlock the write protection

			// Perform the database write operation
			i := 0
			LocalWritesLen := len(LocalWrites)
			log.Printf("Executing %d inserts.\n", LocalWritesLen)

			for i < LocalWritesLen {
				LocalWrites[i](pool.Database)
				i++
			}
		} else {
			pool.WriteLock.Unlock()
			time.Sleep(WRITER_POLL_TIME)
		}
	}
}

func WriteLater(pool *DatabasePool, runnable DatabaseRunnable) {
	pool.WriteLock.Lock()
	pool.PendingWrites = append(pool.PendingWrites, runnable)
	pool.WriteLock.Unlock()
}

func InitDatabasePool(config config.Config) (*DatabasePool, error) {
	db, err := sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			config.DbUserName,
			config.DbPassword,
			config.DbUrl,
			config.DbPort,
			config.DbName,
			config.SslMode))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	db.SetMaxOpenConns(config.DbMaxConnections)
	ret := DatabasePool{Database: db, PendingWrites: make([]DatabaseRunnable, 0)}

	log.Println("Connected to the database successfully. Starting writer thread.")
	go WriterThread(&ret)
	return &ret, nil
}

/*
 * ExecuteOnDatabase will execute code on the database synchronously.
 *
 * @param pool, the database connection pool to use
 * @param runnable the go routine to execute on the database
 */
func ExecuteOnDatabase(pool *DatabasePool, runnable DatabaseRunnable) error {
	return nil
}
