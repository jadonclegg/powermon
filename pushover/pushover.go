package pushover

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"powermon/logging"
	"time"
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
	Priority  int    `json:"priority"`
	Timestamp int64  `json:"timestamp"`
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
			Timestamp: time.Now().Unix(),
			Priority:  1,
		}

		go notification.send()
	}

	return nil
}

func (notification *Notification) send() {
	buff := &bytes.Buffer{}
	enc := json.NewEncoder(buff)
	err := enc.Encode(notification)
	if err != nil {
		return
	}

	count := 0
	// If any error isn't nil, it loops again and tries to send again. When the request succeeds, it will stop the loop
	for {
		// Quit trying to send after 10 tries.
		if count >= 10 {
			logging.Logger.Error("Failed to send Pushover notification 10 times, not trying again.")
			break
		}

		// If trying again, wait 10 seconds
		if count > 0 {
			time.Sleep(time.Second * 10)
		}

		resp, err := http.Post(PushoverURL, ContentType, buff)
		count++
		if err != nil {
			logging.Logger.Error("Failed to send Pushover notification: ", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logging.Logger.Error("Failed to send Pushover notification: ", err)
			}

			logging.Logger.Error(fmt.Sprintf("Failed to send Pushover notification: Non 200 status code: %d, body: %s", resp.StatusCode, body))
			continue
		}

		break
	}
}
