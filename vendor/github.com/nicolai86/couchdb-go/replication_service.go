package couchdb

import "context"

// ReplicationsDatabase is the default replication database name
const ReplicationsDatabase = "_replicator"

// ReplicationService exposes non-admin user management apis
type ReplicationService struct {
	c *Client
}

// UserContext defines execution environment for replications & sessions
type UserContext struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

// Replication contains replication parameters
type Replication struct {
	Document
	Source                 string            `json:"source"`
	Target                 string            `json:"target"`
	Continuous             bool              `json:"continuous"`
	CreateTarget           bool              `json:"create_target"`
	ReplicationID          string            `json:"_replication_id,omitempty"`
	ReplicationState       string            `json:"_replication_state,omitempty"`
	ReplicationStateReason string            `json:"_replication_state_reason,omitempty"`
	Context                *UserContext      `json:"user_ctx,omitempty"`
	Filter                 string            `json:"filter,omitempty"`
	QueryParams            map[string]string `json:"query_params,omitempty"`
}

type ReplicationPayload struct {
	ID           string
	Source       string
	Target       string
	Continuous   bool
	CreateTarget bool
	Filter       string
	QueryParams  map[string]string
	Context      *UserContext
}

func (c *ReplicationService) Create(ctx context.Context, p ReplicationPayload) (*Replication, error) {
	db := c.c.Database(ReplicationsDatabase)
	rep := Replication{
		Document:     Document{ID: p.ID},
		Source:       p.Source,
		Target:       p.Target,
		CreateTarget: p.CreateTarget,
		Continuous:   p.Continuous,
		Filter:       p.Filter,
		QueryParams:  p.QueryParams,
		Context:      p.Context,
	}
	_, err := db.Put(context.Background(), p.ID, rep)
	return &rep, err
}

func (c *ReplicationService) Get(ctx context.Context, id string) (*Replication, error) {
	db := c.c.Database(ReplicationsDatabase)
	rep := Replication{}
	err := db.Get(context.Background(), id, &rep)
	return &rep, err
}

func (c *ReplicationService) Update(ctx context.Context, p ReplicationPayload) (*Replication, error) {
	db := c.c.Database(ReplicationsDatabase)
	rev, err := db.Rev(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	rep := Replication{
		Document: Document{
			ID:  p.ID,
			Rev: rev,
		},
		Source:       p.Source,
		Target:       p.Target,
		CreateTarget: p.CreateTarget,
		Continuous:   p.Continuous,
		Filter:       p.Filter,
		QueryParams:  p.QueryParams,
		Context:      p.Context,
	}
	_, err = db.Put(context.Background(), p.ID, rep)
	return &rep, err
}

func (c *ReplicationService) Delete(ctx context.Context, id string) error {
	db := c.c.Database(ReplicationsDatabase)
	rev, err := db.Rev(ctx, id)
	if err != nil {
		return err
	}
	_, err = db.Delete(ctx, id, rev)
	return err
}
