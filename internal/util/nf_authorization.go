package util

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/openapi/oauth"
	udr_context "github.com/free5gc/udr/internal/context"
)

func AuthorizationCheck(c *gin.Context, serviceName string) error {
	if udr_context.GetSelf().OAuth2Required {
		oauth_err := oauth.VerifyOAuth(c.Request.Header.Get("Authorization"), serviceName,
			udr_context.GetSelf().NrfCertPem)
		if oauth_err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": oauth_err.Error()})
			return oauth_err
		}
	}
	return nil
}
