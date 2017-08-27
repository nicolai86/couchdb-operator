package couchdb

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type ClusterService struct {
	c *Client
}

type ClusterSetup struct {
	Action string `json:"action"`
}

type SetupOptions struct {
	Action         string `json:"action"`
	BindAddress    string `json:"bind_address"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	NodeCount      int    `json:"node_count"`
	Port           int    `json:"port,omitempty"`
	RemoteNode     string `json:"remote_node,omitempty"`
	RemoteUsername string `json:"remote_current_user,omitempty"`
	RemotePassword string `json:"remote_current_password,omitempty"`
}

type AddNodeOptions struct {
	Action   string `json:"action"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *ClusterService) AddNode(opts AddNodeOptions) error {
	opts.Action = "add_node"
	bs, err := json.Marshal(opts)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "/_cluster_setup", bytes.NewBuffer(bs))
	if err != nil {
		return err
	}

	resp, err := s.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	{
		bbs, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Add node resp: %s\n\n", string(bbs))
	}
	return nil
}

func (s *ClusterService) BeginSetup(opts SetupOptions) error {
	opts.Action = "enable_cluster"
	bs, err := json.Marshal(opts)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "/_cluster_setup", bytes.NewBuffer(bs))
	if err != nil {
		return err
	}

	resp, err := s.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	{
		bbs, _ := ioutil.ReadAll(resp.Body)
		log.Printf("begin cluster setup: %s\n\n", string(bbs))
	}
	return nil
}

func (s *ClusterService) EndSetup() error {
	opts := ClusterSetup{"finish_cluster"}
	bs, err := json.Marshal(opts)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "/_cluster_setup", bytes.NewBuffer(bs))
	if err != nil {
		return err
	}

	resp, err := s.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	{
		bbs, _ := ioutil.ReadAll(resp.Body)
		log.Printf("finish cluster setup: %s\n\n", string(bbs))
	}
	return nil
}
