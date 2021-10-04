package services

import (
	"io"

	"github.com/argoproj/notifications-engine/pkg/util/misc"
)

type consoleService struct {
	stdout io.Writer
}

func (c *consoleService) Send(notification Notification, _ Destination) error {
	return misc.PrintFormatted(notification, "yaml", c.stdout)
}

func NewConsoleService(stdout io.Writer) *consoleService {
	return &consoleService{stdout}
}
