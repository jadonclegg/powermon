package commands

import (
	"errors"
	"fmt"
	"os"
	"powermon/logging"
	"powermon/pushover"

	"github.com/sirupsen/logrus"
)

// PowermonCommand holds the settings for the base command Powermon
// These are basically global options, used for both the server and client commands
type PowermonCommand struct {
	// Available commands in powermon
	Client ClientCommand `command:"client" description:"Client mode."`
	Server ServerCommand `command:"server" description:"Server mode."`
	MAC    MACCommand    `command:"mac" description:"List MAC addresses."`

	// Global options
	Verbose bool   `short:"v" long:"verbose" description:"Verbose output"`
	Log     string `short:"l" long:"logfile" description:"Log file"`

	// Pushover config
	PushoverToken string   `short:"k" long:"pushover-token" group:"Pushover" description:"API token to use for pushover notifications."`
	UserTokens    []string `short:"u" long:"user-token" group:"Pushover" description:"User token to send notification to."`
	NickName      string   `short:"n" long:"nickname" group:"Pushover" description:"Nickname in pushover notificatinos."`

	pushoverEnabled bool
}

// Powermon holds the global options, accessible in other files
var Powermon PowermonCommand

// Logger makes it so we don't have to import logging in every command file...
var Logger *logrus.Logger

// Init should be the first thing that any command calls, to initialize the global application options
func (command *PowermonCommand) Init() error {
	logging.Logger = logrus.New()
	Logger = logging.Logger

	// Default output is stdout
	Logger.Out = os.Stdout
	// Use the basic text formatter
	Logger.SetFormatter(&logrus.TextFormatter{})

	// If the verbose flag is enabled, show info logs
	if command.Verbose {
		Logger.Level = logrus.InfoLevel
	} else {
		Logger.Level = logrus.WarnLevel
	}

	// If the log file is specified, open or create the file and set it as the output for Logger
	if command.Log != "" {
		file, err := os.OpenFile(command.Log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return err
		}

		Logger.Out = file
	}

	// If the pushover token or user token options are present, check them for validity and set up the Default notifier in pushover/pushover.go
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

	// Todo: Implement the pushover validation API call to check if the user token / api token is valid before continuing

	pushover.Initialize(command.PushoverToken, command.UserTokens)

	return nil
}

// SendNotification wraps the pushover crap so the call can be one line in the other command files. Only sends if the pushover options are passed in and valid
func SendNotification(message string) error {
	if Powermon.pushoverEnabled {
		if Powermon.NickName != "" {
			message = fmt.Sprintf("%s: %s", Powermon.NickName, message)
		}

		// Use the Default notifier to send messages.
		err := pushover.Default.SendNotification(message)
		if err != nil {
			Logger.Error("Pushover notification failed: ", err)
			return err
		}
	}

	return nil
}
