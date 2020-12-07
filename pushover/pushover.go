package pushover

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// PushoverURL is the url of the pushover api
const PushoverURL = "https://api.pushover.net/1/messages.json"

// ContentType is the default content type for requests
const ContentType = "application/json"

// NotifierSettings is a struct to use that lets you create and send a notification using the same token and user every time.
type NotifierSettings struct {
	APIToken   string
	UserTokens []string
}

// Notification object for notifications
type Notification struct {
	Message   string `json:"message"`
	APIToken  string `json:"token"`
	UserToken string `json:"user"`
}

// Default is the default notifier to use, but must be initialized first
var Default NotifierSettings

// Initialize the default notifier
func Initialize(token string, users []string) {
	Default.APIToken = token
	Default.UserTokens = users
}

// SendNotification creates and sends a notification.
func (notifier *NotifierSettings) SendNotification(message string) error {
	if notifier.APIToken == "" || len(notifier.UserTokens) == 0 {
		return errors.New("notifier not initialized. Check APIToken or User tokens")
	}

	for _, user := range notifier.UserTokens {
		notification := Notification{
			Message:   message,
			UserToken: user,
			APIToken:  notifier.APIToken,
		}

		err := notification.send()
		if err != nil {
			return err
		}
	}

	return nil
}

func (notification *Notification) send() error {
	buff := &bytes.Buffer{}
	enc := json.NewEncoder(buff)
	err := enc.Encode(notification)
	if err != nil {
		return err
	}

	resp, err := http.Post(PushoverURL, ContentType, buff)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("Status code from pushover was not 200. \nStatus code: %d.\nBody: %s", resp.StatusCode, body)
	}

	return nil
}
