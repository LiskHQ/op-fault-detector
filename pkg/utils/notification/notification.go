package notification

import (
	"context"

	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	slack "github.com/LiskHQ/op-fault-detector/pkg/utils/notification/channel"
	"go.uber.org/multierr"
)

// Notification holds information on all the supported channels require to communicate with the channel API.
type Notification struct {
	Slack *slack.Slack
}

// NewNotification will return [Notification] with the initialized channel information.
func NewNotification(ctx context.Context, logger log.Logger, notificationConfig *config.Notification) (*Notification, error) {
	newNotification := &Notification{}

	if notificationConfig.Slack != nil {
		client, err := slack.NewClient(ctx, logger, notificationConfig.Slack)
		if err != nil {
			return nil, err
		} else {
			newNotification.Slack = client
		}
	}

	return newNotification, nil
}

// Notify sends a message to the available channels and returns combined error from different channels if any.
func (n *Notification) Notify(msg string) error {
	var combinedError error

	if n.Slack != nil {
		if err := n.Slack.Notify(msg); err != nil {
			combinedError = multierr.Append(combinedError, err)
		}
	}

	return combinedError
}
