package authentication

import (
	"context"

	"github.com/sirupsen/logrus"
)

type FlashCall struct {
	log       *logrus.Entry
	host      string
	AppID     string
	AppSecret string
}

func New(log *logrus.Logger, host, appID, appSecret string) *FlashCall {
	fc := FlashCall{
		log:       log.WithField("module", "flashcall"),
		host:      host,
		AppID:     appID,
		AppSecret: appSecret,
	}
	return &fc
}

func (fc *FlashCall) Authenticate(ctx context.Context, phone, code string) error {
	fc.log.Infof("phone number %s verified successfully with code %s", phone, code)
	return nil
}

func (fc *FlashCall) VerifyPhone(ctx context.Context, phone string) error {
	fc.log.Infof("calling %s", phone)
	return nil
}
