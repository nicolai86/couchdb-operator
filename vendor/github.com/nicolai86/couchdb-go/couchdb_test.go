// +build !integration

package couchdb

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
)

var (
	client     *Client
	playground *Database
)

type testDoc struct {
	Document
	Name string `json:"name"`
}

func TestMain(m *testing.M) {
	if os.Getenv("COUCHDB_HOST_PORT") == "" {
		fmt.Println("Skipping couchdb tests as COUCHDB_HOST_PORT is not configured")
		os.Exit(0)
	}
	auth := func(c *Client) error { return nil }
	if os.Getenv("COUCHDB_USERNAME") != "" {
		auth = WithBasicAuthentication(os.Getenv("COUCHDB_USERNAME"), os.Getenv("COUCHDB_PASSWORD"))
	}

	c, _ := New(os.Getenv("COUCHDB_HOST_PORT"), &http.Client{}, auth)
	client = c

	func() {
		if exists, _ := client.Databases.Exists("playground"); !exists {
			client.Databases.Create("playground", DatabaseClusterOptions{})
		}
		playground = client.Database("playground")

		playground.Put(context.Background(), "employee:michael", testDoc{
			Name: "Michael",
		})
		playground.Put(context.Background(), "employee:raphael", testDoc{
			Name: "Raphael",
		})
		playground.Put(context.Background(), "pet:yumi", testDoc{
			Name: "Yumi",
		})
	}()

	flag.Parse()
	ret := m.Run()
	if err := client.Databases.Delete("playground"); err != nil {
		log.Fatalf("failed to delete test database: %v\n", err.Error())
	}
	os.Exit(ret)
}
