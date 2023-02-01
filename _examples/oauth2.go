package examples

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/go-session/session"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	clientStore   = store.NewClientStore()
	authConfig    = new(oauth2.Config)
	globalToken   *oauth2.Token
	authServerURL string
)

const (
	usernameConfig = "test"
	passwordConfig = "test"
	userIDConfig   = "123456"
)

// AuthServerHandler is a simple http.Handler that implements go-oauth2 server.
func AuthServerHandler() http.Handler {
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	manager.MustTokenStorage(store.NewMemoryTokenStore())

	manager.MapAccessGenerate(generates.NewAccessGenerate())

	manager.MapClientStorage(clientStore)

	srv := server.NewServer(server.NewConfig(), manager)

	srv.SetPasswordAuthorizationHandler(
		func(ctx context.Context,
			clientID,
			username,
			password string) (userID string, err error) {
			if username == usernameConfig && password == passwordConfig {
				userID = userIDConfig
			}
			return
		},
	)

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
		_ = dumpRequest(os.Stdout, "authorize", r)

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
		_ = dumpRequest(os.Stdout, "token", r)

		err := srv.HandleTokenRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		_ = dumpRequest(os.Stdout, "test", r)

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

// AuthClientHandler is a simple http.Handler that implements go-oauth2 client.
func AuthClientHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u := authConfig.AuthCodeURL("xyz",
			oauth2.SetAuthURLParam("code_challenge", genCodeChallengeS256("s256example")),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"))
		http.Redirect(w, r, u, http.StatusFound)
	})

	mux.HandleFunc("/oauth2", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		state := r.Form.Get("state")
		if state != "xyz" {
			http.Error(w, "State invalid", http.StatusBadRequest)
			return
		}
		code := r.Form.Get("code")
		if code == "" {
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}
		token, err := authConfig.Exchange(context.Background(),
			code,
			oauth2.SetAuthURLParam("code_verifier", "s256example"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		globalToken = token

		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		_ = e.Encode(token)
	})

	mux.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		if globalToken == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		globalToken.Expiry = time.Now()
		token, err := authConfig.TokenSource(context.Background(), globalToken).Token()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		globalToken = token
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		_ = e.Encode(token)
	})

	mux.HandleFunc("/try", func(w http.ResponseWriter, r *http.Request) {
		if globalToken == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		accessTokenURL := fmt.Sprintf("%s/test?access_token=%s",
			authServerURL,
			globalToken.AccessToken)
		resp, err := http.Get(accessTokenURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.Copy(w, resp.Body)
	})

	mux.HandleFunc("/pwd", func(w http.ResponseWriter, r *http.Request) {
		token, err := authConfig.PasswordCredentialsToken(context.Background(),
			usernameConfig,
			passwordConfig,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		globalToken = token
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		_ = e.Encode(token)
	})

	mux.HandleFunc("/client", func(w http.ResponseWriter, r *http.Request) {
		cfg := clientcredentials.Config{
			ClientID:     authConfig.ClientID,
			ClientSecret: authConfig.ClientSecret,
			TokenURL:     authConfig.Endpoint.TokenURL,
		}

		token, err := cfg.Token(context.Background())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		_ = e.Encode(token)
	})

	return mux
}

func dumpRequest(writer io.Writer, header string, r *http.Request) error {
	data, err := httputil.DumpRequest(r, true)
	if err != nil {
		return err
	}
	_, _ = writer.Write([]byte("\n" + header + ": \n"))
	_, _ = writer.Write(data)
	return nil
}

func userAuthorizeHandler(
	w http.ResponseWriter,
	r *http.Request,
) (userID string, err error) {
	_ = dumpRequest(os.Stdout, "userAuthorizeHandler", r)

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
	_ = dumpRequest(os.Stdout, "login", r)

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

	outputHTML(w, r, "static/login.html")
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	_ = dumpRequest(os.Stdout, "auth", r)

	st, err := session.Start(nil, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := st.Get("LoggedInUserID"); !ok {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	outputHTML(w, r, "static/auth.html")
}

func outputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
}

func genCodeChallengeS256(s string) string {
	s256 := sha256.Sum256([]byte(s))
	return base64.URLEncoding.EncodeToString(s256[:])
}
