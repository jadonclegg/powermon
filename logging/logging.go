package logging

import "github.com/sirupsen/logrus"

// Logger used for all logging. Initialized by calling command.Init() in all the command files that need to
var Logger *logrus.Logger
