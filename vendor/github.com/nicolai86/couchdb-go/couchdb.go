// Package couchdb provides a wrapper around the couchdb HTTP API
package couchdb

// ErrorResponse represents a > 400 http api error from couchdb
type ErrorResponse struct {
	Type   string `json:"error"`
	Reason string `json:"reason"`
}

func (e ErrorResponse) Error() string {
	return e.Reason
}

// New returns a configured couchdb client
func New(host string, client HTTPExecutor, configs ...func(*Client) error) (*Client, error) {
	c := &Client{
		Host:   host,
		client: client,
	}
	for _, config := range configs {
		if err := config(c); err != nil {
			return nil, err
		}
	}
	c.Databases = &DatabaseService{c}
	c.Admins = &AdminUserService{c}
	c.Users = &UserService{c}
	c.Replications = &ReplicationService{c}
	c.Sessions = &SessionService{c}
	c.Cluster = &ClusterService{c}
	return c, c.Check()
}
