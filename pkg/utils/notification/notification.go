package notification

import (
	"context"

	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	slack "github.com/LiskHQ/op-fault-detector/pkg/utils/notification/channel"
)

// Notification holds information on all the supported channels require to communicate with the channel API.
type Notification struct {
	slack *slack.Slack
}

// NewNotification will return [Notification] with the initialized channel information.
func NewNotification(ctx context.Context, logger log.Logger, notificationConfig *config.Notification) (*Notification, error) {
	newNotification := &Notification{}

	if notificationConfig.Slack != nil {
		client, err := slack.NewClient(ctx, logger, notificationConfig.Slack)
		if err != nil {
			return nil, err
		} else {
			newNotification.slack = client
		}
	}

	return newNotification, nil
}

// Notify sends a message to the available channels.
func (n *Notification) Notify(msg string) *[]error {
	var errors []error

	if n.slack != nil {
		if err := n.slack.Notify(msg); err != nil {
			errors = append(errors, err)
		}
	}

	return &errors
}
