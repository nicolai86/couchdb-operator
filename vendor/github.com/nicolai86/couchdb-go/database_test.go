// +build !integration

package couchdb

import "testing"

func TestDatabase_NotExisting(t *testing.T) {
	t.Parallel()

	dbName := "foobar"
	exists, err := client.Databases.Exists(dbName)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatalf("Expected database %q to not exist, but does.", dbName)
	}
}

func TestDatabase_Exists(t *testing.T) {
	t.Parallel()

	dbName := "_replicator"
	exists, err := client.Databases.Exists(dbName)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatalf("Expected database %q to exist, but didn't.", dbName)
	}
}

func TestClient_Create(t *testing.T) {
	if err := client.Databases.Create("new-db", DatabaseClusterOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestClient_Delete(t *testing.T) {
	if err := client.Databases.Delete("new-db"); err != nil {
		t.Fatal(err)
	}
}
