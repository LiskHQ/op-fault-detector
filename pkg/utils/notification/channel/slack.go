package slack

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/LiskHQ/op-fault-detector/pkg/config"
	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/slack-go/slack"
)

type SlackClient interface {
	PostMessageContext(ctx context.Context, channelID string, options ...slack.MsgOption) (string, string, error)
}

// Slack holds all necessary information require to communicate with the Slack API.
type Slack struct {
	client    SlackClient
	channelID string
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
		channelID: slackConfig.ChannelID,
		ctx:       ctx,
		logger:    logger,
	}, nil
}

// Notify sends a message to the slack channel.
func (s *Slack) Notify(msg string) error {
	_, timestamp, err := s.client.PostMessageContext(
		s.ctx,
		s.channelID,
		slack.MsgOptionText(msg, false),
	)
	if err != nil {
		s.logger.Errorf("Failed to send notification to the channel %s, error: %v", s.channelID, err)
		return err
	}

	parts := strings.Split(timestamp, ".")

	timeInMS, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return err
	}
	localTime := time.UnixMilli(timeInMS * int64(time.Microsecond)).Local()

	s.logger.Infof("Message successfully sent to the channel %s at %s", s.channelID, localTime.String())
	return nil
}

func GetSlackClient(ctx context.Context, logger log.Logger, client SlackClient, slackConfig *config.SlackConfig) *Slack {
	return &Slack{
		client:    client,
		channelID: slackConfig.ChannelID,
		ctx:       ctx,
		logger:    logger,
	}
}
