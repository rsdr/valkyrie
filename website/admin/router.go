package admin

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"

	radio "github.com/R-a-dio/valkyrie"
	"github.com/R-a-dio/valkyrie/config"
	"github.com/R-a-dio/valkyrie/templates"
	"github.com/R-a-dio/valkyrie/util/secret"
	vmiddleware "github.com/R-a-dio/valkyrie/website/middleware"
	"github.com/R-a-dio/valkyrie/website/shared"
	"github.com/rs/zerolog/hlog"
	"github.com/spf13/afero"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

func NewState(
	_ context.Context,
	cfg config.Config,
	dp secret.Secret,
	songSecret secret.Secret,
	newsCache *shared.NewsCache,
	storage radio.StorageService,
	search radio.SearchService,
	siteTmpl *templates.Site,
	exec templates.Executor,
	sessionManager *scs.SessionManager,
	auth vmiddleware.Authentication,
	fs afero.Fs,
) State {
	return State{
		Config:           cfg,
		Daypass:          dp,
		SongSecret:       songSecret,
		News:             newsCache,
		Storage:          storage,
		Search:           search,
		Templates:        siteTmpl,
		TemplateExecutor: exec,
		SessionManager:   sessionManager,
		Authentication:   auth,
		FS:               fs,
	}
}

type State struct {
	config.Config

	Daypass          secret.Secret
	SongSecret       secret.Secret
	News             *shared.NewsCache
	Storage          radio.StorageService
	Search           radio.SearchService
	Templates        *templates.Site
	TemplateExecutor templates.Executor
	SessionManager   *scs.SessionManager
	Authentication   vmiddleware.Authentication
	FS               afero.Fs
}

func Route(ctx context.Context, s State) func(chi.Router) {
	return func(r chi.Router) {
		// the login middleware will require atleast the active permission
		r = r.With(
			s.Authentication.LoginMiddleware,
		)
		r.Handle("/set-theme", templates.SetThemeHandler(
			templates.ThemeAdminCookieName,
			s.Templates.ResolveThemeName,
		))
		r.HandleFunc("/", s.GetHome)
		r.Get("/profile", s.GetProfile)
		r.Post("/profile", s.PostProfile)
		r.Get("/pending",
			vmiddleware.RequirePermission(radio.PermPendingView, s.GetPending))
		r.Post("/pending",
			vmiddleware.RequirePermission(radio.PermPendingEdit, s.PostPending))
		r.Get("/pending-song/{SubmissionID:[0-9]+}",
			vmiddleware.RequirePermission(radio.PermPendingView, s.GetPendingSong))
		r.Get("/songs",
			vmiddleware.RequirePermission(radio.PermDatabaseView, s.GetSongs))
		r.Post("/songs",
			vmiddleware.RequirePermission(radio.PermDatabaseEdit, s.PostSongs))
		r.Get("/users",
			vmiddleware.RequirePermission(radio.PermAdmin, s.GetUsersList))
		r.Get("/news",
			vmiddleware.RequirePermission(radio.PermNews, s.GetNews))
		r.Get("/news/{NewsID:[0-9]+|new}",
			vmiddleware.RequirePermission(radio.PermNews, s.GetNewsEntry))
		r.Post("/news/{NewsID:[0-9]+|new}",
			vmiddleware.RequirePermission(radio.PermNews, s.PostNewsEntry))
		r.Post("/news/render",
			vmiddleware.RequirePermission(radio.PermNews, s.PostNewsRender))
		r.Get("/queue", // TODO: change permission to queue specific
			vmiddleware.RequirePermission(radio.PermAdmin, s.GetQueue))
		r.Get("/schedule", // TODO: change permission to schedule specific
			vmiddleware.RequirePermission(radio.PermAdmin, s.GetSchedule))
		r.Get("/listeners", // TODO: change permissions to listener specific
			vmiddleware.RequirePermission(radio.PermAdmin, s.GetListeners))

		// proxy to the grafana host
		grafana, _ := url.Parse("http://localhost:3000")
		proxy := httputil.NewSingleHostReverseProxy(grafana)
		r.Handle("/grafana/*", vmiddleware.RequirePermission(radio.PermDev, proxy.ServeHTTP))

		// debug handlers, might not be needed later
		r.Post("/api/streamer/stop", vmiddleware.RequirePermission(radio.PermAdmin, s.PostStreamerStop))
		r.Post("/api/website/reload-templates", vmiddleware.RequirePermission(radio.PermAdmin, s.PostReloadTemplates))
	}
}

func (s *State) PostStreamerStop(w http.ResponseWriter, r *http.Request) {
	s.Conf().Streamer.Client().Stop(r.Context(), false)
}

func (s *State) errorHandler(w http.ResponseWriter, r *http.Request, err error, msg string) {
	// TODO: implement this better
	hlog.FromRequest(r).Error().Err(err).Msg(msg)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
