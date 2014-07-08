# oauth2 [![wercker status](https://app.wercker.com/status/cfc6a7d08ba203b6d40aa0b3bd69b477/s/ "wercker status")](https://app.wercker.com/project/bykey/cfc6a7d08ba203b6d40aa0b3bd69b477)

Allows your Martini application to support user login via an OAuth 2.0 backend. Requires [`sessions`](https://github.com/martini-contrib/sessions) middleware. Google, Facebook, LinkedIn and Github sign-in are currently supported. Once endpoints are provided, this middleware can work with any OAuth 2.0 backend.

## Usage

~~~ go
package main

import (
  "fmt"
  "github.com/go-martini/martini"
  "github.com/martini-contrib/oauth2"
  "github.com/martini-contrib/sessions"
)

func main() {
  m := martini.Classic()
  m.Use(sessions.Sessions("my_session", sessions.NewCookieStore([]byte("secret123"))))
  m.Use(oauth2.Google(&oauth2.Options{
    ClientId:      "client_id",
    ClientSecret:  "client_secret",
    RedirectURL:   "redirect_url",
    Scopes:        []string{"https://www.googleapis.com/auth/drive"},
    LoginCallback: func (tokens oauth2.Tokens){

      fmt.Printf("Login Callback %s", tokens.Access())
    },
  }))

  // Tokens are injected to the handlers
  m.Get("/", func(tokens oauth2.Tokens) string {
    if tokens.IsExpired() {
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
~~~

If a route requires login, you can add `oauth2.LoginRequired` to the handler chain. If user is not logged, they will be automatically redirected to the login path.

~~~ go
m.Get("/login-required", oauth2.LoginRequired, func() ...)
~~~

## Auth flow

* `/login` will redirect user to the OAuth 2.0 provider's permissions dialog. If there is a `next` query param provided, user is redirected to the next page afterwards.
* If user agrees to connect, OAuth 2.0 provider will redirect to `/oauth2callback` to let your app to make the handshake. You need to register `/oauth2callback` as a Redirect URL in your application settings.
* `/logout` will log the user out. If there is a `next` query param provided, user is redirected to the next page afterwards.

You can customize the login, logout, oauth2callback and error paths:

~~~ go
oauth2.PathLogin = "/oauth2login"
oauth2.PathLogout = "/oauth2logout"
...
~~~

## Authors

* [Burcu Dogan](http://github.com/rakyll)
