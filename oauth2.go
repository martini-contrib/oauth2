// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package oauth2 contains Martini handlers to provide
// user login via an OAuth 2.0 backend.
package oauth2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
	"golang.org/x/oauth2"
)

const (
	codeRedirect = 302
	keyToken     = "oauth2_token"
	keyNextPage  = "next"
)

var (
	// PathLogin is the path to handle OAuth 2.0 logins.
	PathLogin = "/login"
	// PathLogout is the path to handle OAuth 2.0 logouts.
	PathLogout = "/logout"
	// PathCallback is the path to handle callback from OAuth 2.0 backend
	// to exchange credentials.
	PathCallback = "/oauth2callback"
	// PathError is the path to handle error cases.
	PathError = "/oauth2error"
)

// Options represents OAuth 2.0 credentials and
// further configuration to be used during access token retrieval.
type Options oauth2.Options

// Tokens represents a container that contains user's OAuth 2.0 access and refresh tokens.
type Tokens interface {
	Access() string
	Refresh() string
	Expired() bool
	ExpiryTime() time.Time
}

type token struct {
	oauth2.Token
}

// Access returns the access token.
func (t *token) Access() string {
	return t.AccessToken
}

// Refresh returns the refresh token.
func (t *token) Refresh() string {
	return t.RefreshToken
}

// Expired returns whether the access token is expired or not.
func (t *token) Expired() bool {
	if t == nil {
		return true
	}
	return t.Token.Expired()
}

// ExpiryTime returns the expiry time of the user's access token.
func (t *token) ExpiryTime() time.Time {
	return t.Expiry
}

// String returns the string representation of the token.
func (t *token) String() string {
	return fmt.Sprintf("tokens: %v", t)
}

// Google returns a new Google OAuth 2.0 backend endpoint.
func Google(opt ...oauth2.Option) martini.Handler {
	return NewOAuth2Provider(append(opt, oauth2.Endpoint(
		"https://accounts.google.com/o/oauth2/auth",
		"https://accounts.google.com/o/oauth2/token"),
	))
}

// Github returns a new Github OAuth 2.0 backend endpoint.
func Github(opt ...oauth2.Option) martini.Handler {
	return NewOAuth2Provider(append(opt, oauth2.Endpoint(
		"https://github.com/login/oauth/authorize",
		"https://github.com/login/oauth/access_token"),
	))
}

func Facebook(opt ...oauth2.Option) martini.Handler {
	return NewOAuth2Provider(append(opt, oauth2.Endpoint(
		"https://www.facebook.com/dialog/oauth",
		"https://graph.facebook.com/oauth/access_token"),
	))
}

func LinkedIn(opt ...oauth2.Option) martini.Handler {
	return NewOAuth2Provider(append(opt, oauth2.Endpoint(
		"https://www.linkedin.com/uas/oauth2/authorization",
		"https://www.linkedin.com/uas/oauth2/accessToken"),
	))
}

// NewOAuth2Provider returns a generic OAuth 2.0 backend endpoint.
func NewOAuth2Provider(opts []oauth2.Option) martini.Handler {
	f, err := oauth2.New(opts...)
	if err != nil {
		// TODO(jbd): Don't panic.
		panic(fmt.Sprintf("oauth2: %s", err))
	}

	return func(s sessions.Session, c martini.Context, w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			switch r.URL.Path {
			case PathLogin:
				login(f, s, w, r)
			case PathLogout:
				logout(s, w, r)
			case PathCallback:
				handleOAuth2Callback(f, s, w, r)
			}
		}
		tk := unmarshallToken(s)
		if tk != nil {
			// check if the access token is expired
			if tk.Expired() && tk.Refresh() == "" {
				s.Delete(keyToken)
				tk = nil
			}
		}
		// Inject tokens.
		c.MapTo(tk, (*Tokens)(nil))
	}
}

// Handler that redirects user to the login page
// if user is not logged in.
// Sample usage:
// m.Get("/login-required", oauth2.LoginRequired, func() ... {})
var LoginRequired martini.Handler = func() martini.Handler {
	return func(s sessions.Session, c martini.Context, w http.ResponseWriter, r *http.Request) {
		token := unmarshallToken(s)
		if token == nil || token.Expired() {
			next := url.QueryEscape(r.URL.RequestURI())
			http.Redirect(w, r, PathLogin+"?next="+next, codeRedirect)
		}
	}
}()

func login(f *oauth2.Options, s sessions.Session, w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get(keyNextPage))
	if s.Get(keyToken) == nil {
		// User is not logged in.
		if next == "" {
			next = "/"
		}
		http.Redirect(w, r, f.AuthCodeURL(next, "", ""), codeRedirect)
		return
	}
	// No need to login, redirect to the next page.
	http.Redirect(w, r, next, codeRedirect)
}

func logout(s sessions.Session, w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get(keyNextPage))
	s.Delete(keyToken)
	http.Redirect(w, r, next, codeRedirect)
}

func handleOAuth2Callback(f *oauth2.Options, s sessions.Session, w http.ResponseWriter, r *http.Request) {
	next := extractPath(r.URL.Query().Get("state"))
	code := r.URL.Query().Get("code")
	t, err := f.NewTransportFromCode(code)
	if err != nil {
		// Pass the error message, or allow dev to provide its own
		// error handler.
		http.Redirect(w, r, PathError, codeRedirect)
		return
	}
	// Store the credentials in the session.
	val, _ := json.Marshal(t.Token())
	s.Set(keyToken, val)
	http.Redirect(w, r, next, codeRedirect)
}

func unmarshallToken(s sessions.Session) (t *token) {
	if s.Get(keyToken) == nil {
		return
	}
	data := s.Get(keyToken).([]byte)
	var tk oauth2.Token
	json.Unmarshal(data, &tk)
	return &token{tk}
}

func extractPath(next string) string {
	n, err := url.Parse(next)
	if err != nil {
		return "/"
	}
	return n.Path
}
