package examples

import (
	"net/http"

	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
)

var (
	ClientID     = "aaa"
	ClientSecret = "bbb"
)

func OAuth2Handler() http.Handler {
	tokenStore, err := store.NewMemoryTokenStore()
	if err != nil {
		panic(err)
	}

	clientStore := store.NewClientStore()

	err = clientStore.Set(ClientID, &models.Client{
		ID:     ClientID,
		Secret: ClientSecret,
	})
	if err != nil {
		panic(err)
	}

	mng := manage.NewDefaultManager()

	mng.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	mng.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)

	mng.MapTokenStorage(tokenStore)
	mng.MapClientStorage(clientStore)

	srv := server.NewDefaultServer(mng)

	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	mux := http.NewServeMux()

	mux.HandleFunc("/token",
		func(w http.ResponseWriter, r *http.Request) {
			_ = srv.HandleTokenRequest(w, r)
		})

	mux.HandleFunc("/protected",
		validateToken(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("protected_content"))
		}, srv))

	return mux
}

func validateToken(f http.HandlerFunc, srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := srv.ValidationBearerToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		f.ServeHTTP(w, r)
	}
}
