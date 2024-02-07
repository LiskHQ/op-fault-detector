package notification

import (
	"context"
	"os"

	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/slack-go/slack"
)

// Slack holds all necessary information require to communicate with the Slack API.
type Slack struct {
	client    *slack.Client
	ChannelID string
	ctx       context.Context
	logger    log.Logger
}

// NewSlack will return [Slack] with the initialized configuration.
func NewSlack(ctx context.Context, logger log.Logger, slackConfig *config.SlackConfig) *Slack {
	slackAccessToken := os.Getenv("SLACK_ACCESS_TOKEN_KEY")
	client := slack.New(slackAccessToken)

	return &Slack{
		client:    client,
		ChannelID: slackConfig.ChannelID,
		ctx:       ctx,
		logger:    logger,
	}
}

// SendNotification sends a message to the slack channel.
func (s *Slack) SendNotification(msg string) error {
	_, timestamp, err := s.client.PostMessageContext(
		s.ctx,
		s.ChannelID,
		slack.MsgOptionText(msg, false),
	)

	if err != nil {
		s.logger.Errorf("Error while sending notification to the channel %s, error: %w", s.ChannelID, err)
		return err
	}

	s.logger.Infof("Message successfully sent to the channel %s at %s", s.ChannelID, timestamp)
	return nil
}
