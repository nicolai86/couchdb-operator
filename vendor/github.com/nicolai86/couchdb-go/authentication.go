package couchdb

import (
	"net/http"
)

// Authentication is used to allow couchdb to support multiple authentication methods
type Authentication interface {
	Decorate(*http.Request) error
}

// BasicAuthentication uses basic authorization for couchdb API requests
type BasicAuthentication struct {
	username string
	password string
}

// Decorate adds basic auth headers to the request
func (b BasicAuthentication) Decorate(r *http.Request) error {
	r.SetBasicAuth(b.username, b.password)
	return nil
}

// WithBasicAuthentication returns a new basic authentication mechanism
func WithBasicAuthentication(username, password string) func(*Client) error {
	authenticator := BasicAuthentication{
		username: username,
		password: password,
	}
	return func(c *Client) error {
		c.Authenticator = authenticator
		return nil
	}
}
