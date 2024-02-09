package slack

import (
	"context"
	"fmt"
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

// NewClient will return [Slack] with the initialized configuration.
func NewClient(ctx context.Context, logger log.Logger, slackConfig *config.SlackConfig) (*Slack, error) {
	var slackAccessToken string
	if slackAccessToken = os.Getenv("SLACK_ACCESS_TOKEN_KEY"); len(slackAccessToken) == 0 {
		return nil, fmt.Errorf("failed to access slack API token from the environment")
	}

	client := slack.New(slackAccessToken)

	return &Slack{
		client:    client,
		ChannelID: slackConfig.ChannelID,
		ctx:       ctx,
		logger:    logger,
	}, nil
}

// Notify sends a message to the slack channel.
func (s *Slack) Notify(msg string) error {
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