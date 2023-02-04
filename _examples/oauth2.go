package examples

import (
	"net/http"

	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
)

var (
	clientStore = store.NewClientStore()
)

// OAuth2Handler is a simple http.Handler that implements go-oauth2 server.
func OAuth2Handler() http.Handler {
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	manager.MustTokenStorage(store.NewMemoryTokenStore())

	manager.MapClientStorage(clientStore)

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)
	manager.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = srv.HandleTokenRequest(w, r)
	})

	mux.HandleFunc("/protected", validateToken(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Protected!!!"))
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
