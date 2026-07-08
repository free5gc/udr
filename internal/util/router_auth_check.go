package util

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"


	udr_context "github.com/free5gc/udr/internal/context"
	"github.com/free5gc/udr/internal/logger"
)

const RequesterNfInstanceIDCtxKey = "requesterNfInstanceId"

type oauth2Requirement interface {
	OAuth2Enabled() bool
}

type requesterClaims struct {
	NfInstanceID string
}

func parseRequesterClaims(authorization string) (requesterClaims, bool) {
	fields := strings.Fields(authorization)
	if len(fields) < 2 {
		return requesterClaims{}, false
	}

	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(fields[1], claims)
	if err != nil {
		return requesterClaims{}, false
	}

	requester := requesterClaims{}
	if sub, ok := claims["sub"].(string); ok {
		requester.NfInstanceID = sub
	}

	return requester, requester.NfInstanceID != ""
}

type RouterAuthorizationCheck struct {
	serviceName models.ServiceName
}

func NewRouterAuthorizationCheck(serviceName models.ServiceName) *RouterAuthorizationCheck {
	return &RouterAuthorizationCheck{
		serviceName: serviceName,
	}
}

func (rac *RouterAuthorizationCheck) Check(c *gin.Context, udrContext udr_context.NFContext) {
	token := c.Request.Header.Get("Authorization")
	err := udrContext.AuthorizationCheck(token, rac.serviceName)
	if err != nil {
		logger.UtilLog.Debugf("RouterAuthorizationCheck: Check Unauthorized: %s", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	if requirement, ok := udrContext.(oauth2Requirement); ok && !requirement.OAuth2Enabled() {
		logger.UtilLog.Debugf("RouterAuthorizationCheck: OAuth2 disabled, skip requester claims parsing")
		logger.UtilLog.Debugf("RouterAuthorizationCheck: Check Authorized")
		return
	}

	if requester, ok := parseRequesterClaims(token); ok {
		if requester.NfInstanceID != "" {
			c.Set(RequesterNfInstanceIDCtxKey, requester.NfInstanceID)
		}
	}

	logger.UtilLog.Debugf("RouterAuthorizationCheck: Check Authorized")
}
