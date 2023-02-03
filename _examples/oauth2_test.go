package examples

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestOAuth2(t *testing.T) {
	server := httptest.NewServer(OAuth2Handler())
	defer server.Close()

	clientID := uuid.New().String()[:8]
	clientSecret := uuid.New().String()[:8]

	err := clientStore.Set(clientID, &models.Client{
		ID:     clientID,
		Secret: clientSecret,
		Domain: server.URL,
	})
	assert.NoError(t, err)

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: server.URL + "/token",
		},
	}

	token := &oauth2.Token{
		AccessToken:  uuid.New().String()[:8],
		RefreshToken: uuid.New().String()[:8],
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

	e.GET("/protected").Expect().Status(http.StatusBadRequest)

	rr := e.GET("/token").WithQueryObject(map[string]interface{}{
		"grant_type":    "client_credentials",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"scope":         "all",
	}).Expect()

	rr.Status(http.StatusOK)

	accessToken := rr.JSON().Path("$.access_token").String().Raw()
	if accessToken != token.AccessToken {
		token.AccessToken = accessToken
	}

	e.GET("/protected").Expect().Status(http.StatusOK)
}
