package admin

import (
	"context"
	"net/http"
	"time"

	radio "github.com/R-a-dio/valkyrie"
	"github.com/R-a-dio/valkyrie/config"
	"github.com/R-a-dio/valkyrie/templates"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi"
)

type State struct {
	config.Config

	Storage   radio.StorageService
	Templates templates.Templates
}

func newSessionManager() *scs.SessionManager {
	s := scs.New()
	s.Lifetime = 150 * 24 * time.Hour
	s.Cookie = scs.SessionCookie{
		Name: "admin",
		//SameSite: http.SameSiteStrictMode,
		Secure: true,
	}
	return s
}

func Router(ctx context.Context, s State) chi.Router {
	sessionManager := scs.New()
	sessionManager.Store = NewSessionStore(ctx, s.Storage)
	sessionManager.Codec = JSONCodec{}
	sessionManager.Lifetime = 150 * 24 * time.Hour
	sessionManager.Cookie = scs.SessionCookie{
		Name: "admin",
		//SameSite: http.SameSiteStrictMode,
		Secure: true,
	}

	authentication := NewAuthentication(s.Storage, s.Templates, sessionManager)

	r := chi.NewRouter()
	r.Use(sessionManager.LoadAndSave)
	r.Get("/logout", authentication.LogoutHandler)
	adminRouter := chi.NewRouter()
	adminRouter.Use(authentication.LoginMiddleware)
	adminRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	r.Mount("/", adminRouter)
	return r
}
