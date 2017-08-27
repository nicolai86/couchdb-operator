package couchdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// UsersDatabase is the default authentication database name
const UsersDatabase = "_users"

// UserService exposes non-admin user management apis
type UserService struct {
	c *Client
}

// CreateUserPayload defines all parameters required when creating a regular user
type CreateUserPayload struct {
	Name     string
	Password string
	Roles    []string
}

// User contains all information for interacting with couchdb user documents
type User struct {
	Document
	Name     string   `json:"name"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
	Type     string   `json:"type"`
}

// Create adds a new user to couchdb
func (c *UserService) Create(ctx context.Context, p CreateUserPayload) (*User, error) {
	user := User{
		Document: Document{
			ID: fmt.Sprintf("org.couchdb.user:%s", p.Name),
		},
		Name:     p.Name,
		Password: p.Password,
		Roles:    p.Roles,
		Type:     "user",
	}
	db := c.c.Database(UsersDatabase)
	rev, err := db.Put(ctx, user.ID, user)
	if err != nil {
		return nil, err
	}
	user.Rev = rev
	return &user, nil
}

// UpdateUserPayload defiens all parameters for updating existing users
type UpdateUserPayload struct {
	ID       string
	Name     string
	Password string
	Roles    []string
}

// Update modifies an existing user inside couchdb
func (c *UserService) Update(ctx context.Context, p UpdateUserPayload) (*User, error) {
	db := c.c.Database(UsersDatabase)
	rev, err := db.Rev(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	user := User{
		Document: Document{
			ID:  p.ID,
			Rev: rev,
		},
		Name:     p.Name,
		Password: p.Password,
		Roles:    p.Roles,
		Type:     "user",
	}
	rev, err = db.Put(ctx, p.ID, user)
	if err != nil {
		return nil, err
	}
	user.Rev = rev
	return &user, nil
}

// Delete removes a regular couchdb user
func (c *UserService) Delete(ctx context.Context, id string) error {
	db := c.c.Database(UsersDatabase)
	rev, err := db.Rev(ctx, id)
	if err != nil {
		return err
	}
	_, err = db.Delete(ctx, id, rev)
	return err
}

// Get fetches a regular couchdb user
func (c *UserService) Get(ctx context.Context, id string) (*User, error) {
	db := c.c.Database(UsersDatabase)
	user := User{}
	if err := db.Get(ctx, id, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// AdminUserService exposes administrative user management
type AdminUserService struct {
	c *Client
}

// ClusterOptions allows the user to target different nodes in the cluster
type ClusterOptions struct {
	Node string
}

// Create adds a new administrative user
func (c *AdminUserService) Create(ctx context.Context, name, password string, opts ClusterOptions) error {
	path := fmt.Sprintf("/_config/admins/%s", name)
	if c.c.CouchDB.HasClusterSupport() {
		path = fmt.Sprintf("/_node/%s/_config/admins/%s", opts.Node, name)
	}
	req, err := http.NewRequest("PUT", path, strings.NewReader(fmt.Sprintf("%q", password)))
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	_, err = c.c.Do(req)
	if err != nil {
		return err
	}
	return nil
}

// Update modifies an existimg administrative user
func (c *AdminUserService) Update(ctx context.Context, name, password string, opts ClusterOptions) error {
	return c.Create(ctx, name, password, opts)
}

// List fetches all administrative users
func (c *AdminUserService) List(ctx context.Context, opts ClusterOptions) ([]string, error) {
	path := "/_config/admins"
	if c.c.CouchDB.HasClusterSupport() {
		path = fmt.Sprintf("/_node/%s/_config/admins", opts.Node)
	}
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data := map[string]string{}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bs, &data); err != nil {
		return nil, err
	}
	users := []string{}
	for name := range data {
		users = append(users, name)
	}
	return users, nil
}

// Delete removes an administrative user
func (c *AdminUserService) Delete(ctx context.Context, name string, opts ClusterOptions) error {
	path := fmt.Sprintf("/_config/admins/%s", name)
	if c.c.CouchDB.HasClusterSupport() {
		path = fmt.Sprintf("/_node/%s/_config/admins/%s", opts.Node, name)
	}
	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	_, err = c.c.Do(req)
	if err != nil {
		return err
	}
	return nil
}
