package notification

import (
	"context"

	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	slack "github.com/LiskHQ/op-fault-detector/pkg/utils/notification/channel"
)

// Channel holds information on all the supported channels require to communicate with the channel API.
type Channel struct {
	slack *slack.Slack
}

// NewClient will return [Channel] with the initialized channel information.
func NewClient(ctx context.Context, logger log.Logger, notificationConfig *config.Notification) (*Channel, error) {
	var slackClient *slack.Slack
	if notificationConfig.Slack != nil {
		client, err := slack.NewClient(ctx, logger, notificationConfig.Slack)
		if err != nil {
			return nil, err
		} else {
			slackClient = client
		}
	}

	return &Channel{
		slack: slackClient,
	}, nil
}

// Notify sends a message to the available channels.
func (c *Channel) Notify(msg string) *[]error {
	var errors []error

	if c.slack != nil {
		if err := c.slack.Notify(msg); err != nil {
			errors = append(errors, err)
		}
	}

	return &errors
}
