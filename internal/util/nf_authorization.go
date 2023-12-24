package util

import (
	"github.com/gin-gonic/gin"

	"github.com/free5gc/openapi/oauth"
	udr_context "github.com/free5gc/udr/internal/context"
)

func AuthorizationCheck(c *gin.Context, serviceName string) error {
	if udr_context.GetSelf().OAuth2Required {
		return oauth.VerifyOAuth(c.Request.Header.Get("Authorization"), serviceName,
			udr_context.GetSelf().NrfCertPem)
	}
	return nil
}
