// +build !integration

package couchdb

import (
	"context"
	"testing"
)

type playgroundResults struct {
	Results
	Rows []struct {
		ID    string `json:"id"`
		Value struct {
			Document
			Name string
		} `json:"doc"`
	} `json:"rows"`
}

func TestDatabase_AllDocs(t *testing.T) {
	t.Parallel()

	t.Run("include_docs", func(t *testing.T) {
		var results playgroundResults
		if err := playground.AllDocs(context.Background(), AllDocOpts{
			IncludeDocs: true,
			StartKey:    "employee:",
			EndKey:      "employee:{}",
		}, &results); err != nil {
			t.Fatal(err)
		}

		for _, doc := range results.Rows {
			if doc.Value.Name == "" {
				t.Fatal("Expected doc to be included")
			}
		}
	})

	t.Run("startkey,endkey", func(t *testing.T) {
		var results playgroundResults
		if err := playground.AllDocs(context.Background(), AllDocOpts{
			StartKey: "employee:",
			EndKey:   "employee:{}",
		}, &results); err != nil {
			t.Fatal(err)
		}
		if len(results.Rows) != 2 {
			t.Fatalf("Expected 2 rows, got %d", len(results.Rows))
		}
	})

	t.Run("limit", func(t *testing.T) {
		var results playgroundResults
		if err := playground.AllDocs(context.Background(), AllDocOpts{
			Limit: 2,
		}, &results); err != nil {
			t.Fatal(err)
		}
		if len(results.Rows) != 2 {
			t.Fatalf("Expected 2 rows, got %d", len(results.Rows))
		}
	})

	t.Run("simple", func(t *testing.T) {
		var results playgroundResults
		if err := playground.AllDocs(context.Background(), AllDocOpts{}, &results); err != nil {
			t.Fatal(err)
		}

		if len(results.Rows) != 4 {
			t.Fatalf("Expected 4 rows, got %d", len(results.Rows))
		}

		expected := []string{
			"pet:yumi",
			"employee:raphael",
			"employee:michael",
			"_design/company",
		}
		for _, id := range expected {
			known := false
			for _, d := range results.Rows {
				known = known || d.ID == id
			}
			if !known {
				t.Fatalf("Expected %q to be known, but wasn't", id)
			}
		}
	})
}

func TestDatabase_Put(t *testing.T) {
	t.Parallel()

	var doc = testDoc{
		Document: Document{
			ID: "employee:martin",
		},
		Name: "Martin",
	}

	db := client.Database("put-test")
	client.Databases.Create(db.Name, DatabaseClusterOptions{})
	defer client.Databases.Delete(db.Name)

	t.Run("insert", func(t *testing.T) {
		rev, err := db.Put(context.Background(), doc.ID, &doc)
		if err != nil {
			t.Fatal(err)
		}
		if rev == "" {
			t.Fatal("Expected to receive a document revision, but got nothing")
		}
		doc.Rev = rev
	})

	t.Run("update", func(t *testing.T) {
		doc.Name = "Klaus"
		rev, err := db.Put(context.Background(), doc.ID, &doc)
		if err != nil {
			t.Fatal(err)
		}
		if rev == doc.Rev {
			t.Fatalf("Expected update to succeed, but didn't")
		}
	})
}

func TestDatabase_Delete(t *testing.T) {
	t.Parallel()

	db := client.Database("delete-test")
	client.Databases.Create(db.Name, DatabaseClusterOptions{})
	defer client.Databases.Delete(db.Name)

	rev, _ := db.Put(context.Background(), "test", Document{
		ID: "test",
	})
	rmRev, err := db.Delete(context.Background(), "test", rev)
	if err != nil {
		t.Fatal(err)
	}
	if rmRev == rev {
		t.Fatalf("Expected new revision, but got %q", rmRev)
	}
}

func TestDatabase_Get(t *testing.T) {
	t.Parallel()

	db := client.Database("_users")
	t.Run("known", func(t *testing.T) {
		t.Parallel()

		docID := "_design/_auth"
		var doc Document
		err := db.Get(context.Background(), docID, &doc)
		if err != nil {
			t.Fatal(err)
		}
		if doc.ID != docID {
			t.Fatalf("Expected doc %q, but got %q", docID, doc.ID)
		}
	})
	t.Run("unknown", func(t *testing.T) {
		t.Parallel()

		docID := "admin"
		var doc Document

		if err := db.Get(context.Background(), docID, &doc); err == nil {
			t.Fatal(err)
		}
	})
}

func TestDatabase_Rev(t *testing.T) {
	t.Parallel()
	db := client.Database("_users")

	t.Run("known", func(t *testing.T) {
		t.Parallel()

		docID := "_design/_auth"
		var doc Document
		err := db.Get(context.Background(), docID, &doc)
		if err != nil {
			t.Fatal(err)
		}

		rev, err := db.Rev(context.Background(), docID)
		if err != nil {
			t.Fatal(err)
		}
		if rev != doc.Rev {
			t.Fatalf("Expected revisions to match, but didn't: %q != %q", rev, doc.Rev)
		}
	})
	t.Run("unknown", func(t *testing.T) {
		t.Parallel()

		db := client.Database("_users")
		if _, err := db.Rev(context.Background(), "admin"); err == nil {
			t.Fatal(err)
		}
	})
}
