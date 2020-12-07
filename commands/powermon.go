package commands

import (
	"errors"
	"fmt"
	"os"
	"powermon/pushover"

	"github.com/sirupsen/logrus"
)

// PowermonCommand holds the settings for the base command Powermon
type PowermonCommand struct {
	Client ClientCommand `command:"client" description:"Client mode."`
	Server ServerCommand `command:"server" description:"Server mode."`
	MAC    MACCommand    `command:"mac" description:"List MAC addresses."`

	Verbose bool   `short:"v" long:"verbose" description:"Verbose output"`
	Log     string `short:"l" long:"logfile" description:"Log file"`

	PushoverToken string   `short:"k" long:"pushover-token" group:"Pushover" description:"API token to use for pushover notifications."`
	UserTokens    []string `short:"u" long:"user-token" group:"Pushover" description:"User token to send notification to."`
	NickName      string   `short:"n" long:"nickname" group:"Pushover" description:"Nickname in pushover notificatinos."`

	pushoverEnabled bool
}

// Powermon holds the global options
var Powermon PowermonCommand

// Logger used for all logging
var Logger *logrus.Logger

// Init runs stuff needed for all commands, like initializing logging
func (command *PowermonCommand) Init() error {
	Logger = logrus.New()
	Logger.Out = os.Stdout
	Logger.SetFormatter(&logrus.TextFormatter{})

	if command.Verbose {
		Logger.Level = logrus.InfoLevel
	} else {
		Logger.Level = logrus.WarnLevel
	}

	if command.Log != "" {
		file, err := os.OpenFile(command.Log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return err
		}

		Logger.Out = file
	}

	if command.PushoverToken != "" || len(command.UserTokens) > 0 {
		if err := command.initializePushover(); err != nil {
			return err
		}

		command.pushoverEnabled = true
	}
	return nil
}

func (command *PowermonCommand) initializePushover() error {
	if command.PushoverToken == "" {
		return errors.New("API Token must be specified using -k [--pushover-token] flag")
	}

	if len(command.UserTokens) == 0 {
		return errors.New("Must specify one or more user tokens to send notifications to using -u [--user-token] flag")
	}

	pushover.Initialize(command.PushoverToken, command.UserTokens)

	return nil
}

// SendNotification wraps the pushover crap so the call can be one line. Only sends if enabled
func SendNotification(message string) error {
	if Powermon.pushoverEnabled {
		if Powermon.NickName != "" {
			message = fmt.Sprintf("%s: %s", Powermon.NickName, message)
		}

		err := pushover.Default.SendNotification(message)
		if err != nil {
			Logger.Error("Pushover notification failed: ", err)
			return err
		}
	}

	return nil
}
