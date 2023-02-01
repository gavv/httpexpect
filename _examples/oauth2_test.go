package examples

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-oauth2/oauth2/v4/models"
)

var (
	cookieJar = httpexpect.NewCookieJar()
)

func withServerConfig(t *testing.T, serverURL string) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL: serverURL,
		Client: &http.Client{
			Jar: cookieJar,
		},
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})
}

func withClientConfig(t *testing.T, clientURL string) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL: clientURL,
		Client: &http.Client{
			Jar: cookieJar,
		},
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})
}

func TestOAuth2(t *testing.T) {
	server := httptest.NewServer(AuthServerHandler())
	defer server.Close()

	authServerURL = server.URL

	client := httptest.NewServer(AuthClientHandler())
	defer client.Close()

	var (
		clientID     = "CLIENT_ID"
		clientSecret = "CLIENT_SECRET"
	)

	_ = clientStore.Set(clientID, &models.Client{
		ID:     clientID,
		Secret: clientSecret,
		Domain: client.URL,
	})

	authConfig.ClientID = clientID
	authConfig.ClientSecret = clientSecret
	authConfig.Scopes = []string{"all"}
	authConfig.RedirectURL = client.URL + "/oauth2"
	authConfig.Endpoint.AuthURL = server.URL + "/oauth/authorize"
	authConfig.Endpoint.TokenURL = server.URL + "/oauth/token"

	withClientConfig(t, client.URL).GET("/").Expect().
		Status(http.StatusOK).
		Body().IsEqual(
		func() string {
			f, _ := os.Open("./static/login.html")
			b, _ := io.ReadAll(f)
			return string(b)
		}())

	withServerConfig(t, server.URL).POST("/login").
		WithForm(map[string]string{
			"username": usernameConfig,
			"password": passwordConfig,
		}).Expect().Status(http.StatusOK).
		Body().IsEqual(
		func() string {
			f, _ := os.Open("./static/auth.html")
			b, _ := io.ReadAll(f)
			return string(b)
		}())

	oauthResp := withServerConfig(t, server.URL).POST("/oauth/authorize").
		Expect().Status(http.StatusOK).JSON()
	oauthResp.Path("$.access_token").String().NotEmpty()
	oauthResp.Path("$.token_type").String().IsEqual("Bearer")
	oauthResp.Path("$.refresh_token").String().NotEmpty()
	oauthResp.Path("$.expiry").String().NotEmpty()

	tryResp := withClientConfig(t, client.URL).GET("/try").Expect().
		Status(http.StatusOK).JSON()
	tryResp.Path("$.client_id").IsEqual(clientID)
	tryResp.Path("$.expires_in").IsNumber()
	tryResp.Path("$.user_id").IsEqual(usernameConfig)
}
