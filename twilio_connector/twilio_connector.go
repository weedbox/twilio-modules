package twilio_connector

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

var logger *zap.Logger

const (
	DefaultAccountSID = "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	DefaultAuthToken  = "f2xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
)

type SendSMSReq struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Msg   string `json:"msg"`
}

type TwilioConnector struct {
	params Params
	logger *zap.Logger
	client *twilio.RestClient
	scope  string
}

type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    *zap.Logger
}

func Module(scope string) fx.Option {

	var m *TwilioConnector

	return fx.Module(
		scope,
		fx.Provide(func(p Params) *TwilioConnector {

			logger = p.Logger.Named(scope)

			m := &TwilioConnector{
				params: p,
				logger: logger,
				scope:  scope,
			}

			m.initDefaultConfigs()

			return m
		}),
		fx.Populate(&m),
		fx.Invoke(func(p Params) *TwilioConnector {

			p.Lifecycle.Append(
				fx.Hook{
					OnStart: m.onStart,
					OnStop:  m.onStop,
				},
			)

			return m
		}),
	)
}

func (c *TwilioConnector) getConfigPath(key string) string {
	return fmt.Sprintf("%s.%s", c.scope, key)
}

func (c *TwilioConnector) initDefaultConfigs() {
	viper.SetDefault(c.getConfigPath("account_sid"), DefaultAccountSID)
	viper.SetDefault(c.getConfigPath("auth_token"), DefaultAuthToken)
}

func (c *TwilioConnector) onStart(ctx context.Context) error {

	sid := viper.GetString(c.getConfigPath("account_sid"))
	token := viper.GetString(c.getConfigPath("auth_token"))

	logger.Info("Starting TwilioConnector",
		zap.String("account_sid", sid),
		zap.String("token", token),
	)

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: sid,
		Password: token,
	})

	c.client = client

	return nil
}

func (c *TwilioConnector) onStop(ctx context.Context) error {

	c.logger.Info("Stopped TwilioConnector")

	return nil
}

func (c *TwilioConnector) SendSMS(req *SendSMSReq) error {
	params := &openapi.CreateMessageParams{}
	params.SetFrom(req.From)
	params.SetTo(req.To)
	params.SetBody(req.Msg)

	resp, err := c.client.Api.CreateMessage(params)
	if err != nil {
		return err
	}

	c.logger.Info(fmt.Sprintf("Sid: %s", *resp.Sid))
	return nil
}

func (c *TwilioConnector) GetClient() *twilio.RestClient {
	return c.client
}
