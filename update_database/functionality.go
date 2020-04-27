package main

import (
	"github.com/getlantern/errors"
	"sync"
)

var updateLock = sync.RWMutex{}

func updateLen() int {
	updateLock.RLock()
	i := db.Len
	updateLock.RUnlock()
	return i
}

func updatePush(Item map[string]interface{}) error {
	// Get the update hash.
	UpdateHash, ok := Item["update_hash"].(string)
	if !ok {
		return errors.New("update hash needs to be set as a string")
	}

	// Lock the updates.
	updateLock.Lock()

	// Push the item to the updates.
	db.Updates = append(db.Updates, Item)

	// Set the key index to the length.
	db.KeyIndex[UpdateHash] = db.Len

	// Add 1 to length.
	db.Len++

	// Write the DB to disk.
	db.Write()

	// Unlock the updates.
	updateLock.Unlock()

	// Return no errors.
	return nil
}

func updateRm(UpdateHash string) error {
	// Lock the DB.
	updateLock.Lock()

	// Check the update exists.
	index, ok := db.KeyIndex[UpdateHash]
	if !ok {
		updateLock.Unlock()
		return errors.New("update not found")
	}

	// Loop through all of the indexes.
	for i := index; i < db.Len; i++ {
		// Get the item from the arrays key.
		Key := db.Updates[i]["update_hash"].(string)

		// Update the keys from the index to subtract 1.
		db.KeyIndex[Key] -= 1
	}

	// Remove this item by its index.
	delete(db.KeyIndex, UpdateHash)
	copy(db.Updates[index:], db.Updates[index+1:])
	db.Updates = db.Updates[:len(db.Updates)-1]

	// Remove 1 from the length.
	db.Len -= 1

	// Write the updates to disk.
	db.Write()

	// Unlock the updates.
	updateLock.Unlock()

	// Return no errors.
	return nil
}

func updatesBefore(UpdateHash string, Params map[string]interface{}) ([]map[string]interface{}, error) {
	// Read lock the DB.
	updateLock.RLock()

	// Check the update exists.
	index, ok := db.KeyIndex[UpdateHash]
	if !ok {
		updateLock.RUnlock()
		return nil, errors.New("update not found")
	}

	// Make an array the length of index.
	a := make([]map[string]interface{}, 0, index)

	// Get all before this index.
	for i := 0; i < index; i++ {
		item := db.Updates[i]
		b := false
		for k, v := range Params {
			if v != item[k] {
				b = true
				break
			}
		}
		if b {
			continue
		}
		a = append(a, item)
	}

	// Read unlock the DB.
	updateLock.RUnlock()

	// Return the updates.
	return a, nil
}

func updatesAfter(UpdateHash string, Params map[string]interface{}) ([]map[string]interface{}, error) {
	// Read lock the DB.
	updateLock.RLock()

	// Check the update exists.
	index, ok := db.KeyIndex[UpdateHash]
	if !ok {
		updateLock.RUnlock()
		return nil, errors.New("update not found")
	}

	// Make an array the length of index.
	a := make([]map[string]interface{}, 0, index)

	// Get all before this index.
	for i := index + 1; i < db.Len; i++ {
		item := db.Updates[i]
		b := false
		for k, v := range Params {
			if v != item[k] {
				b = true
				break
			}
		}
		if b {
			continue
		}
		a = append(a, item)
	}

	// Read unlock the DB.
	updateLock.RUnlock()

	// Return the updates.
	return a, nil
}

func updateInfo(UpdateHash string) (map[string]interface{}, error) {
	// Read lock the DB.
	updateLock.RLock()

	// Check the update exists.
	index, ok := db.KeyIndex[UpdateHash]
	if !ok {
		updateLock.RUnlock()
		return nil, errors.New("update not found")
	}

	// Get the item by the index.
	x := db.Updates[index]

	// Read unlock the DB.
	updateLock.RUnlock()

	// Return the updates.
	return x, nil
}

func latestUpdate() (string, error) {
	// Read lock the DB.
	updateLock.RLock()

	// Get the DB len.
	length := db.Len

	// Return a error if length is 0.
	if length == 0 {
		return "", errors.New("database length is 0")
	}

	// Get the version hash.
	Hash := db.Updates[length-1]["update_hash"].(string)

	// Read unlock the DB.
	updateLock.RUnlock()

	// Return the hash.
	return Hash, nil
}
