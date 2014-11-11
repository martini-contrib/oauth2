package oauth2_test

import (
	"testing"

	"github.com/go-martini/martini"
	goauth2 "github.com/golang/oauth2"
	"github.com/martini-contrib/oauth2"
	"github.com/martini-contrib/sessions"
)

// TODO(jbd): Remove after Go 1.4.
// Related to https://codereview.appspot.com/107320046
func TestA(t *testing.T) {}

func ExampleLogin() {
	m := martini.Classic()
	m.Use(sessions.Sessions("my_session", sessions.NewCookieStore([]byte("secret123"))))
	m.Use(oauth2.Google(
		goauth2.Client("client_id", "client_secret"),
		goauth2.RedirectURL("redirect_url"),
		goauth2.Scope("https://www.googleapis.com/auth/drive"),
	))
	// Tokens are injected to the handlers
	m.Get("/", func(tokens oauth2.Tokens) string {
		if tokens.Expired() {
			return "not logged in, or the access token is expired"
		}
		return "logged in"
	})

	// Routes that require a logged in user
	// can be protected with oauth2.LoginRequired handler.
	// If the user is not authenticated, they will be
	// redirected to the login path.
	m.Get("/restrict", oauth2.LoginRequired, func(tokens oauth2.Tokens) string {
		return tokens.Access()
	})

	m.Run()
}
