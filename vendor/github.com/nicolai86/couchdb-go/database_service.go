package couchdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// DatabaseService exposes database management apis
type DatabaseService struct {
	c *Client
}

// DefaultReplicaCount defines the default database replication count
const DefaultReplicaCount = 3

// DefaultShardCount defines the default database sharding count
const DefaultShardCount = 8

// DatabaseClusterOptions controls replication and sharding configuration for couchdb 2.x
type DatabaseClusterOptions struct {
	Replicas int
	Shards   int
}

// Create creates a new database by calling PUT /{db}
func (d *DatabaseService) Create(name string, opts DatabaseClusterOptions) error {
	req, err := http.NewRequest("PUT", fmt.Sprintf("/%s", name), nil)
	if err != nil {
		return err
	}
	if d.c.CouchDB.HasClusterSupport() {
		vs := url.Values{}
		replica := DefaultReplicaCount
		if opts.Replicas != 0 {
			replica = opts.Replicas
		}
		shards := DefaultShardCount
		if opts.Shards != 0 {
			shards = opts.Shards
		}
		vs.Set("n", strconv.Itoa(replica))
		vs.Set("q", strconv.Itoa(shards))
		req.URL.RawQuery = vs.Encode()
	}

	_, err = d.c.Do(req)
	return err
}

// Delete removes a database
func (d *DatabaseService) Delete(name string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/%s", name), nil)
	if err != nil {
		return err
	}

	_, err = d.c.Do(req)
	return err
}

// DatabaseMeta contains ever changing meta data about a single database
type DatabaseMeta struct {
	Name                  string `json:"db_name"`
	DocumentCount         int    `json:"doc_count"`
	DocumentDeletionCount int    `json:"doc_del_count"`
	CompactRunning        bool   `json:"compact_running"`
	DiskSize              int    `json:"disk_size"`
	DataSize              int    `json:"data_size"`
	InstanceStartTime     string `json:"instance_start_time"`
	DiskFormatVersion     int    `json:"disk_format_version"`
}

// Meta looks up database metadata
func (d *DatabaseService) Meta(name string) (DatabaseMeta, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/%s", name), nil)
	if err != nil {
		return DatabaseMeta{}, err
	}
	resp, err := d.c.Do(req)
	if err != nil {
		return DatabaseMeta{}, err
	}
	defer resp.Body.Close()
	meta := DatabaseMeta{}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return DatabaseMeta{}, err
	}
	err = json.Unmarshal(bs, &meta)
	return meta, err
}

// Exists checks if the given database exists with a HEAD /{db} request
func (d *DatabaseService) Exists(name string) (bool, error) {
	req, err := http.NewRequest("HEAD", fmt.Sprintf("/%s", name), nil)
	if err != nil {
		return false, err
	}
	resp, err := d.c.Do(req)
	if err != nil {
		return false, err
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}
