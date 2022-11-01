package security

import (
	"container/list"
	"errors"
	"log"
	"math/rand"
	"sync"
	"time"
)

const MAX_TRIES = 10
const MAX_REMOVALS = 2500
const MAX_REMOVAL_TIME = 15 * 1000 * time.Millisecond
const MAXIMUM_NONCES = 25000
const NONCE_POLL_TIME = 100 * time.Millisecond

/**
 * The nonces are 64 bit integers, the map here will map them to their time of creation.
 */
type NonceManager struct {
	NonceMap          map[int64]time.Time
	NonceRemovalQueue *list.List
	NonceLock         sync.Mutex
}

/**
 * Initialises the nonce manager and starts the polling thread
 */
func (n *NonceManager) InitNonceManager() {
	log.Println("Set randomised seed")
	rand.Seed(time.Now().Unix())

	// Init the struct
	n.NonceLock.Lock()
	n.NonceMap = make(map[int64]time.Time)
	n.NonceRemovalQueue = list.New()
	n.NonceLock.Unlock()

	// Start polling thread
	go func() {
		log.Println("Started nonce manager poling thread")
		defer log.Fatal("The nonce manger thread stopped")

		for true {
			n.NonceLock.Lock()
			// Attempt removals
			for i := 0; i < MAX_REMOVALS && n.NonceRemovalQueue.Len() > 0; i++ {
				el := n.NonceRemovalQueue.Front()
				key := el.Value.(int64)
				addedTime, inMap := n.NonceMap[key]

				if !inMap {
					n.NonceRemovalQueue.Remove(el)
				} else if time.Now().Sub(addedTime) >= MAX_REMOVAL_TIME {
					n.NonceRemovalQueue.Remove(el)
					delete(n.NonceMap, key)
				}
			}
			n.NonceLock.Unlock()

			time.Sleep(NONCE_POLL_TIME)
		}
	}()
}

/**
 * Gets a nonce, returning an error if the maximum amount of nonces is met.
 *
 * @return int64 the nonce, only use if error is nil
 * @return error an error if the maximum nonce number has been met
 */
func (n *NonceManager) GetNonce() (int64, error) {
	var err error = nil
	var ret int64 = -1

	n.NonceLock.Lock()
	if n.NonceRemovalQueue.Len() >= MAXIMUM_NONCES {
		err = errors.New("Maximum nonce count has been met")
	} else {
		cont := true
		for i := 0; i < MAX_TRIES && cont; i++ {
			ret = rand.Int63()

			// Check that the nonce is not in use already
			_, inMap := n.NonceMap[ret]
			if !inMap {
				cont = false
				n.NonceMap[ret] = time.Now()
				n.NonceRemovalQueue.PushBack(ret)
			}
		}

		if cont {
			err = errors.New("Could not create nonce due to max retries")
		}
	}
	n.NonceLock.Unlock()
	return ret, err
}

/**
 * Tries to use a nonce returing a nice error if it is not valid
 *
 * @return error a nice error message if the nonce cannot be used
 */
func (n *NonceManager) UseNonce(nonce int64) error {
	var err error = nil

	n.NonceLock.Lock()
	addedTime, inMap := n.NonceMap[nonce]
	if !inMap {
		err = errors.New("The nonce is not in the map")
	} else if time.Now().Sub(addedTime) >= MAX_REMOVAL_TIME {
		err = errors.New("The nonce has expired")
	} else {
		// We are gucci ganag and, our nonce is dappa
		delete(n.NonceMap, nonce)
	}
	n.NonceLock.Unlock()

	return err
}
