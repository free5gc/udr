package consumer

import (
	"context"

	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/openapi/oauth"
	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
)

func GetTokenCtx(scope, targetNF string) (context.Context, *models.ProblemDetails, error) {
	if udr_context.GetSelf().OAuth2Required {
		logger.ConsumerLog.Debugln("GetToekenCtx")
		udrSelf := udr_context.GetSelf()
		tok, pd, err := oauth.SendAccTokenReq(udrSelf.NfId, models.NfType_UDR, scope, targetNF, udrSelf.NrfUri)
		if err != nil {
			return nil, pd, err
		}
		return context.WithValue(context.Background(),
			openapi.ContextOAuth2, tok), pd, nil
	}
	return context.TODO(), nil, nil
}
