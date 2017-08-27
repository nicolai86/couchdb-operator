package couchdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// ErrNotFound is a reusable and checkable 404 error
var ErrNotFound = errors.New("Given document ID was not found in couchDB")

// Document contains basic document identifications
type Document struct {
	ID      string `json:"_id,omitempty"`
	Rev     string `json:"_rev,omitempty"`
	Deleted *bool  `json:"_deleted,omitempty"`
}

func revision(etag string) string {
	if etag == "" {
		return ""
	}
	return etag[1 : len(etag)-1]
}

// AllDocOpts defines parameters which can be passed to APIs returning multiple documents
type AllDocOpts struct {
	Skip        int
	Limit       int
	IncludeDocs bool
	StartKey    string
	EndKey      string
}

func (d *Database) bulkGet(ctx context.Context, path string, opts AllDocOpts, results interface{}) error {
	req, _ := http.NewRequest("GET", path, nil)
	req = req.WithContext(ctx)

	values := req.URL.Query()
	if opts.Limit == 0 {
		opts.Limit = 100
	}
	values.Set("skip", strconv.Itoa(opts.Skip))
	values.Set("limit", strconv.Itoa(opts.Limit))
	values.Set("include_docs", strconv.FormatBool(opts.IncludeDocs))
	if opts.StartKey != "" {
		values.Set("startkey", fmt.Sprintf("%q", opts.StartKey))
	}
	if opts.EndKey != "" {
		values.Set("endkey", fmt.Sprintf("%q", opts.EndKey))
	}
	req.URL.RawQuery = values.Encode()

	resp, err := d.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("couchdb: GET %s returned %d", req.URL.Path, resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &results); err != nil {
		return err
	}

	return nil
}

// AllDocs fetches all documents from couchdb
func (d *Database) AllDocs(ctx context.Context, opts AllDocOpts, results interface{}) error {
	return d.bulkGet(ctx, "/_all_docs", opts, results)
}

// DocumentReadWriter is the interface that groups the basic Read and Write methods.
type DocumentReadWriter interface {
	DocumentReader
	DocumentWriter
}

// DocumentReader abstracts read access to a specific database
type DocumentReader interface {
	Get(context.Context, string, interface{}) error
}

// Get fetches a document identified by it's id. GET /{db}/{id}
// this results in couchdb automatically returning the latest revision of the document
//
//  var doc couchdb.Document
//  db.Get("some-id", &doc)
func (d *Database) Get(ctx context.Context, id string, doc interface{}) error {
	req, _ := http.NewRequest("GET", fmt.Sprintf("/%s", id), nil)
	req = req.WithContext(ctx)
	resp, err := d.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return ErrNotFound
		}

		return fmt.Errorf("couchdb: GET %s returned %d", id, resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}
	return nil
}

// DocumentWriter abstracts write access to a specific database
type DocumentWriter interface {
	Put(context.Context, string, interface{}) (string, error)
}

// Put creates or updates a document, returning the new revision. PUT /{db}/{id}
//
//  var doc = couchdb.Document{
//    ID: "whatever",
//    Rev: "1-62bc3c4d01e43ee9d0cead8cd7c76041",
//  }
//  rev, err := db.Put(doc.ID, &doc)
//  // … modify doc
//  doc.Rev = rev
//  db.Put(doc.ID, &doc)
func (d *Database) Put(ctx context.Context, id string, doc interface{}) (string, error) {
	bs, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/%s", id), bytes.NewReader(bs))
	req = req.WithContext(ctx)
	resp, err := d.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("couchdb: PUT %s returned %d", id, resp.StatusCode)
	}
	return revision(resp.Header.Get("Etag")), nil
}

// Delete removes a document from a database
func (d *Database) Delete(ctx context.Context, id, rev string) (string, error) {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%s", id), nil)
	req = req.WithContext(ctx)
	values := req.URL.Query()
	values.Set("rev", rev)
	req.URL.RawQuery = values.Encode()

	resp, err := d.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("couchdb: DELETE %s returned %d", id, resp.StatusCode)
	}
	return revision(resp.Header.Get("Etag")), nil
}

// Rev fetches the latest revision for a document. HEAD /{db}/{id}
func (d *Database) Rev(ctx context.Context, id string) (string, error) {
	req, _ := http.NewRequest("HEAD", fmt.Sprintf("/%s", id), nil)
	req = req.WithContext(ctx)
	resp, err := d.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("couchdb: HEAD %s returned %d", id, resp.StatusCode)
	}
	return revision(resp.Header.Get("Etag")), nil
}
