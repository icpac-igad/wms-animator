package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

type APIError interface {
	// APIError returns an HTTP status code and an API-safe error message.
	APIError() (int, string)
}
type appError struct {
	Message string
	Status  int
}

func (e appError) Error() string {
	return e.Message
}

func (e appError) APIError() (int, string) {
	return e.Status, e.Message
}

func JSONError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)

	w.Header().Set("Content-Type", "application/json")

	result, err := json.Marshal(struct {
		Message string `json:"message"`
	}{
		Message: message,
	})

	if err != nil {
		w.Write([]byte(fmt.Sprintf("problem marshaling error: %v", message)))
	} else {
		w.Write(result)
	}
}

func JSONHandleError(w http.ResponseWriter, err error) {
	var apiErr APIError
	if errors.As(err, &apiErr) {
		status, msg := apiErr.APIError()
		JSONError(w, status, msg)
	} else {
		JSONError(w, http.StatusInternalServerError, "internal error")
	}
}

// FatalAfter aborts by logging a fatal message, after a time delay.
// The abort can be cancelled by closing the returned channel
func FatalAfter(delaySec int, msg string) chan struct{} {
	chanCancel := make(chan struct{})
	go func() {
		select {
		case <-chanCancel:
			// do nothing if cancelled
			return
		case <-time.After(time.Duration(delaySec) * time.Second):
			// terminate with extreme predjudice
			log.Fatalln(msg)
		}
	}()
	return chanCancel
}
