package couchdb

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// ReplicationService exposes non-admin user management apis
type SessionService struct {
	c *Client
}

type SessionInfo struct {
	Database      string   `json:"authentication_db"`
	Handlers      []string `json:"authentication_handlers"`
	Authenticated string   `json:"authenticated"`
}

type Session struct {
	OK      bool         `json:"ok"`
	Context *UserContext `json:"userCtx"`
	Info    SessionInfo  `json:"info"`
}

func (s *SessionService) Get(ctx context.Context) (*Session, error) {
	req, err := http.NewRequest("GET", "/_session", nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp, err := s.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	sess := Session{}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bs, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}
