package admin

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cybersaksham/gogo/auth"
	"github.com/cybersaksham/gogo/sessions"
)

// AuthViewConfig configures admin authentication views.
type AuthViewConfig struct {
	Site         *Site
	UserStore    auth.UserStore
	SessionStore sessions.Store
	Cookie       sessions.CookieOptions
}

// LoginView authenticates staff users and creates an admin session.
func LoginView(config AuthViewConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("admin login"))
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		user, ok, err := auth.Authenticate(r.Context(), config.UserStore, auth.Credentials{
			Username: r.FormValue("username"),
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		})
		if err != nil {
			http.Error(w, "login failed", http.StatusUnauthorized)
			return
		}
		if !ok || !user.IsStaff {
			http.Error(w, "admin access denied", http.StatusForbidden)
			return
		}
		if err := saveAdminSession(r.Context(), w, config, user); err != nil {
			http.Error(w, "session failed", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, safeNextURL(config.site(), r.FormValue("next")), http.StatusFound)
	})
}

// LogoutView clears the admin session and redirects to login.
func LogoutView(config AuthViewConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		options := normalizeAdminCookie(config.Cookie)
		if config.SessionStore != nil {
			if cookie, err := r.Cookie(options.Name); err == nil && cookie.Value != "" {
				_ = config.SessionStore.Delete(r.Context(), cookie.Value)
			}
		}
		expireAdminCookie(w, options)
		http.Redirect(w, r, config.site().URLPrefix+"/login/", http.StatusFound)
	})
}

// PasswordChangeView changes the current staff user's password.
func PasswordChangeView(config AuthViewConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := currentAdminUser(r, config)
		if !ok || !user.IsStaff || !user.IsActive {
			http.Error(w, "admin access denied", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("password change"))
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		form := auth.PasswordChangeForm{User: user, OldPassword: r.FormValue("old_password"), NewPassword: r.FormValue("new_password")}
		valid, err := form.Validate()
		if err != nil || !valid {
			http.Error(w, "password change failed", http.StatusBadRequest)
			return
		}
		if err := form.Save(); err != nil {
			http.Error(w, "password change failed", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, config.site().URLPrefix+"/password_change/done/", http.StatusFound)
	})
}

func currentAdminUser(r *http.Request, config AuthViewConfig) (auth.User, bool) {
	if user, ok := auth.UserFromContext(r.Context()); ok && user.IsAuthenticated() && !user.IsAnonymous() {
		return user, true
	}
	return SessionPermissionPolicy{
		UserStore:    config.UserStore,
		SessionStore: config.SessionStore,
		Cookie:       config.Cookie,
	}.UserForRequest(r)
}

// PasswordChangeDoneView renders a small completion response.
func PasswordChangeDoneView() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("password changed"))
	})
}

func saveAdminSession(ctx context.Context, w http.ResponseWriter, config AuthViewConfig, user auth.User) error {
	if config.SessionStore == nil {
		return nil
	}
	options := normalizeAdminCookie(config.Cookie)
	session := sessions.NewSession(time.Duration(options.MaxAge) * time.Second)
	session.Set("user_id", strconv.FormatInt(user.ID, 10))
	if err := config.SessionStore.Save(ctx, session); err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     options.Name,
		Value:    session.Key,
		Path:     options.Path,
		Domain:   options.Domain,
		HttpOnly: options.HttpOnly,
		Secure:   options.Secure,
		SameSite: options.SameSite,
		MaxAge:   options.MaxAge,
	}
	if !session.ExpireDate.IsZero() {
		cookie.Expires = session.ExpireDate
	}
	http.SetCookie(w, cookie)
	return nil
}

func expireAdminCookie(w http.ResponseWriter, options sessions.CookieOptions) {
	http.SetCookie(w, &http.Cookie{
		Name:     options.Name,
		Value:    "",
		Path:     options.Path,
		Domain:   options.Domain,
		HttpOnly: options.HttpOnly,
		Secure:   options.Secure,
		SameSite: options.SameSite,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0).UTC(),
	})
}

func normalizeAdminCookie(options sessions.CookieOptions) sessions.CookieOptions {
	if options.Name == "" {
		options.Name = "gogo_sessionid"
	}
	if options.Path == "" {
		options.Path = "/"
	}
	if !options.HttpOnly {
		options.HttpOnly = true
	}
	if options.SameSite == 0 {
		options.SameSite = http.SameSiteLaxMode
	}
	if options.MaxAge == 0 {
		options.MaxAge = 12 * 60 * 60
	}
	return options
}

func (c AuthViewConfig) site() *Site {
	if c.Site != nil {
		return c.Site
	}
	return DefaultSite()
}

func safeNextURL(site *Site, next string) string {
	next = strings.TrimSpace(next)
	if next == "" {
		return site.URLPrefix + "/"
	}
	parsed, err := url.Parse(next)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || strings.HasPrefix(next, "//") || !strings.HasPrefix(next, "/") {
		return site.URLPrefix + "/"
	}
	return next
}
