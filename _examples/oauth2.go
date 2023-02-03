package examples

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/go-session/session"
	"golang.org/x/oauth2"
)

var (
	clientStore = store.NewClientStore()
	authConfig  = new(oauth2.Config)
)

const (
	usernameConfig = "test"
	passwordConfig = "test"
	userIDConfig   = "123456"
)

// OAuth2Handler is a simple http.Handler that implements go-oauth2 server.
func OAuth2Handler() http.Handler {
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	manager.MustTokenStorage(store.NewMemoryTokenStore())

	manager.MapAccessGenerate(generates.NewAccessGenerate())

	manager.MapClientStorage(clientStore)

	srv := server.NewServer(server.NewConfig(), manager)

	srv.SetPasswordAuthorizationHandler(setPasswordAuthHandler)

	srv.SetUserAuthorizationHandler(userAuthorizeHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	mux := http.NewServeMux()

	mux.HandleFunc("/login", loginHandler)

	mux.HandleFunc("/auth", authHandler)

	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		st, err := session.Start(r.Context(), w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var form url.Values
		if v, ok := st.Get("ReturnUri"); ok {
			form = v.(url.Values)
		}
		r.Form = form

		st.Delete("ReturnUri")
		_ = st.Save()

		err = srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		err := srv.HandleTokenRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		token, err := srv.ValidationBearerToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data := map[string]interface{}{
			"expires_in": int64(token.GetAccessCreateAt().
				Add(token.GetAccessExpiresIn()).
				Sub(time.Now()).
				Seconds()),
			"client_id": token.GetClientID(),
			"user_id":   token.GetUserID(),
		}
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		_ = e.Encode(data)
	})

	return mux
}

func setPasswordAuthHandler(_ context.Context, _, username, password string) (userID string, err error) {
	if username == usernameConfig && password == passwordConfig {
		userID = userIDConfig
	}
	return
}

func userAuthorizeHandler(
	w http.ResponseWriter,
	r *http.Request,
) (userID string, err error) {
	st, err := session.Start(r.Context(), w, r)
	if err != nil {
		return
	}

	uid, ok := st.Get("LoggedInUserID")
	if !ok {
		if r.Form == nil {
			_ = r.ParseForm()
		}

		st.Set("ReturnUri", r.Form)
		_ = st.Save()

		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	userID = uid.(string)
	st.Delete("LoggedInUserID")
	_ = st.Save()
	return
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	st, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		if r.Form == nil {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if r.Form.Get("username") != usernameConfig &&
			r.Form.Get("password") != passwordConfig {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		st.Set("LoggedInUserID", r.Form.Get("username"))
		_ = st.Save()

		w.Header().Set("Location", "/auth")
		w.WriteHeader(http.StatusFound)
		return
	}
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	st, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := st.Get("LoggedInUserID"); !ok {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}
}
