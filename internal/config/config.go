package config

import "github.com/kelseyhightower/envconfig"

const (
	MockBroadcasterID    = "12345"
	MockBroadcasterLogin = "mockstreamer"
	MockBroadcasterName  = "MockStreamer"
	MockBotID            = "67890"
	MockBotLogin         = "mockbot"
	MockBotName          = "MockBot"
	MockAppToken         = "mock-app-token"
	MockBotToken         = "mock-bot-token"
	MockUserToken        = "mock-user-token"

	MockKickAppToken      = "mock-kick-app-token"
	MockKickUserToken     = "mock-kick-user-token"
	MockKickBroadcasterID = 12345
	MockKickBotID         = 67890
	MockKickCategoryID    = 1
	MockKickCategoryName  = "Just Chatting"
	MockKickChannelSlug   = "mockstreamer"
)

type Config struct {
	HTTPAddr    string `envconfig:"HTTP_ADDR" default:":7777"`
	WSAddr      string `envconfig:"WS_ADDR" default:":8081"`
	AdminAddr   string `envconfig:"ADMIN_ADDR" default:":3333"`
	SiteBaseURL string `envconfig:"SITE_BASE_URL" default:"http://localhost:3005"`
}

func New() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
