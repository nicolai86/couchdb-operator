package couchdb

import (
	"context"
	"fmt"
	"net/http"
)

// Database is a client for a specific couchdb server & database
type Database struct {
	c    *Client
	Name string
}

// Do forwards requests to the http client, prefixing the URL path with the database name
func (d *Database) Do(req *http.Request) (*http.Response, error) {
	req.URL.Path = fmt.Sprintf("/%s%s", d.Name, req.URL.EscapedPath())
	return d.c.Do(req)
}

type AuthorizationRules struct {
	Names []string `json:"names"`
	Roles []string `json:"roles"`
}

type DatabaseSecurity struct {
	Admins  AuthorizationRules `json:"admins"`
	Members AuthorizationRules `json:"members"`
}

type securityDocument struct {
	Document
	DatabaseSecurity
}

func (db *Database) GetSecurity(ctx context.Context) (*DatabaseSecurity, error) {
	sec := securityDocument{}
	err := db.Get(ctx, "_security", &sec)
	return &sec.DatabaseSecurity, err
}

func (db *Database) SetSecurity(ctx context.Context, sec DatabaseSecurity) error {
	_, err := db.Put(ctx, "_security", securityDocument{
		Document:         Document{ID: "_security"},
		DatabaseSecurity: sec,
	})
	return err
}
