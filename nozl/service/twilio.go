package service

import (
	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/shared"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

func (s *Service) SendMsgTwilio(msg *eventstream.Message) error {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: s.AccountSID,
		Password: s.AuthToken,
	})

	params := &twilioApi.CreateMessageParams{}
	params.SetBody(msg.ReqBody.Payload)
	params.SetFrom(shared.SenderPhoneNumber)
	params.SetTo(msg.ReqBody.Destination)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		shared.Logger.Error(err.Error())
		return err
	}

	shared.Logger.Info(*resp.Sid)

	return nil
}
