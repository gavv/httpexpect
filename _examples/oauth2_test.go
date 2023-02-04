package examples

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"golang.org/x/oauth2"
)

var (
	AccessToken  = "xxx"
	RefreshToken = "yyy"
)

func TestOAuth2(t *testing.T) {
	server := httptest.NewServer(OAuth2Handler())
	defer server.Close()

	config := oauth2.Config{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: server.URL + "/token",
		},
	}

	token := &oauth2.Token{
		AccessToken:  AccessToken,
		RefreshToken: RefreshToken,
		Expiry:       time.Now().Add(5 * time.Minute),
	}

	client := config.Client(context.Background(), token)

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Client:   client,
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})

	e.GET("/protected").
		Expect().
		Status(http.StatusBadRequest)

	tokenResp := e.GET("/token").
		WithQueryObject(map[string]interface{}{
			"grant_type":    "client_credentials",
			"client_id":     ClientID,
			"client_secret": ClientSecret,
			"scope":         "all",
		}).
		Expect()

	tokenResp.Status(http.StatusOK)
	tokenResp.JSON().Path("$.scope").String().IsEqual("all")
	tokenResp.JSON().Path("$.token_type").String().IsEqual("Bearer")
	tokenResp.JSON().Path("$.expires_in").Number().Gt(0)

	accessToken := tokenResp.JSON().Path("$.access_token").String().Raw()
	if token.AccessToken != accessToken {
		token.AccessToken = accessToken
	}

	e.GET("/protected").
		Expect().
		Status(http.StatusOK).
		Body().
		IsEqual("protected_content")
}
