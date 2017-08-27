package couchdb

import (
	"context"
	"fmt"
)

// Results is a struct meant to be embedded in a couchdb request struct with correct
// rows, e.g.
//
//   type UserResults struct {
//       couchdb.Results
//       Users []user `json:"rows"`
//   }
type Results struct {
	Offset    int `json:"offset"`
	TotalRows int `json:"total_rows"`
}

// View defines map & reduce functions for a single view
type View struct {
	MapFn    string `json:"map,omitempty"`
	ReduceFn string `json:"reduce,omitempty"`
}

// DesignDocument describes a language and all associated views
type DesignDocument struct {
	Document
	Language string          `json:"language"`
	Views    map[string]View `json:"views"`
}

// Results executes a request against a couchdb view
func (d *Database) Results(ctx context.Context, design, view string, opts AllDocOpts, results interface{}) error {
	return d.bulkGet(ctx, fmt.Sprintf("/_design/%s/_view/%s", design, view), opts, results)
}
