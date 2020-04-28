package main

import (
	"encoding/json"
	"errors"
	"github.com/jakemakesstuff/structuredhttp"
	"net/url"
	"os"
)

var updateServerHost = "http://127.0.0.1:7000"

func init() {
	e := os.Getenv("UPDATE_SERVER_HOST")
	if e != "" {
		updateServerHost = e
	}
}

// UpdateChunk defines a chunk for the update.
type UpdateChunk struct {
	URL string `json:"url"`
	Hash string `json:"hash"`
}

// Update defines the update model.
type Update struct {
	UpdateHash string `json:"update_hash"`
	Channel string `json:"channel"`
	Chunks []UpdateChunk `json:"chunks"`
	Version string `json:"version"`
	Changelogs string `json:"changelogs"`
	Name string `json:"name"`
}

// PushUpdate is used to push a update to the update database.
func (u *Update) PushUpdate() error {
	r, err := structuredhttp.POST(updateServerHost+"/push").JSON(u).Run()
	if err != nil {
		return err
	}
	if r.RaiseForStatus() != nil {
		text, _ := r.Text()
		return errors.New(text)
	}
	return nil
}

// RemoveUpdate is used to remove a update from the update database.
func (u *Update) RemoveUpdate() error {
	r, err := structuredhttp.GET(updateServerHost+"/rm/"+url.PathEscape(u.UpdateHash)).JSON(u).Run()
	if err != nil {
		return err
	}
	if r.RaiseForStatus() != nil {
		text, _ := r.Text()
		return errors.New(text)
	}
	return nil
}

// GetUpdatesBeforeAfter gets the updates either before or after a update hash.
func GetUpdatesBeforeAfter(Before bool, UpdateHash string, Channels []string) ([]Update, error) {
	// Marshal the channels into JSON. The error is commented out because it cannot fail (the datatype will marshal).
	b, _ := json.Marshal(&Channels)

	// Get what endpoint to use.
	Endpoint := "/after"
	if Before {
		Endpoint = "/before"
	}

	// Create the request.
	r, err := structuredhttp.GET(updateServerHost+Endpoint+"/"+url.PathEscape(UpdateHash)).Query("channel", string(b)).Run()
	if err != nil {
		return nil, err
	}

	// Check the status.
	if r.RaiseForStatus() != nil {
		t, _ := r.Text()
		return nil, errors.New(t)
	}

	// Create the update array.
	var u []Update
	err = r.JSONToPointer(&u)
	if err != nil {
		return nil, err
	}

	// return the updates.
	return u, nil
}

// GetUpdatesLen is used to get the length of the updates database.
func GetUpdatesLen() (int, error) {
	// Create the request.
	r, err := structuredhttp.GET(updateServerHost+"/len").Run()
	if err != nil {
		return 0, err
	}

	// Check the status.
	if r.RaiseForStatus() != nil {
		t, _ := r.Text()
		return 0, errors.New(t)
	}

	// Create the int.
	var i int
	err = r.JSONToPointer(&i)
	if err != nil {
		return 0, err
	}

	// return the int.
	return i, nil
}

// GetLatestHash is used to get the latest update hash.
func GetLatestHash() (string, error) {
	// Create the request.
	r, err := structuredhttp.GET(updateServerHost+"/latest").Run()
	if err != nil {
		return "", err
	}

	// Check the status.
	if r.RaiseForStatus() != nil {
		t, _ := r.Text()
		return "", errors.New(t)
	}

	// Create the string.
	var i string
	err = r.JSONToPointer(&i)
	if err != nil {
		return "", err
	}

	// return the string.
	return "", nil
}

// GetUpdateInfo is used to get a update from a hash.
func GetUpdateInfo(UpdateHash string) (*Update, error) {
	// Create the request.
	r, err := structuredhttp.GET(updateServerHost+"/info/"+url.PathEscape(UpdateHash)).Run()
	if err != nil {
		return nil, err
	}

	// Check the status.
	if r.RaiseForStatus() != nil {
		t, _ := r.Text()
		return nil, errors.New(t)
	}

	// Create the update object.
	var u Update
	err = r.JSONToPointer(&u)
	if err != nil {
		return nil, err
	}

	// return the update object.
	return &u, nil
}
