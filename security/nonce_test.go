package security

import (
	"log"
	"testing"
	"time"
)

var manager NonceManager

func TestInitNonceManager(t *testing.T) {
	manager.InitNonceManager()

	// Check that the manager is unlocked
	manager.NonceLock.Lock()
	manager.NonceLock.Unlock()

	// Wait a while for the polling thread to do its things
	log.Printf("Waiting %d..", NONCE_POLL_TIME)
	time.Sleep(NONCE_POLL_TIME)
	log.Println("DONE")

	// Check that the manager is unlocked
	manager.NonceLock.Lock()
	manager.NonceLock.Unlock()
}

func TestUseNonce(t *testing.T) {
	err := manager.UseNonce(123123)
	if err == nil {
		t.Fail()
	}
}

func TestGetNonce(t *testing.T) {
	for i := 0; i < MAXIMUM_NONCES; i++ {
		nonce, err := manager.GetNonce()
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		err = manager.UseNonce(nonce)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
	}

	// Assert that the max nonce count can be hit
	var terr error = nil
	for i := 0; i < 2*MAXIMUM_NONCES; i++ {
		_, err := manager.GetNonce()
		if err != nil {
			terr = err
			break
		}
	}

	if terr == nil {
		t.Log("Maximum nonce count was not met")
		t.Fail()
	}

	manager.NonceLock.Lock()
	size := manager.NonceRemovalQueue.Len()

	if size != MAXIMUM_NONCES {
		t.Log("The nonce removal queue should be full")
		t.Log(size)
		t.Fail()
	}

	// Change all the times to a time in the past
	for k, v := range manager.NonceMap {
		manager.NonceMap[k] = v.Add(-2 * MAX_REMOVAL_TIME)
	}
	manager.NonceLock.Unlock()

	// Wait a while for the polling thread to do its things
	log.Printf("Waiting %d..", NONCE_POLL_TIME)
	time.Sleep(NONCE_POLL_TIME + 2000*time.Millisecond)
	log.Println("DONE")

	manager.NonceLock.Lock()
	size = manager.NonceRemovalQueue.Len()
	manager.NonceLock.Unlock()

	if size != 0 {
		t.Log("There should be no nonces in the queue at this point")
		t.Log(size)
		t.Fail()
	}
}
