package couchdb

import (
	"context"
	"testing"
)

func TestAdminUserService(t *testing.T) {
	memberships, err := client.Membership()
	opts := ClusterOptions{}
	if err == nil {
		opts.Node = memberships.AllNodes[0]
	}

	t.Run("Create", func(t *testing.T) {
		err := client.Admins.Create(context.Background(), "test", "test", opts)
		if err != nil {
			t.Fatal(err.Error())
		}
	})

	t.Run("List", func(t *testing.T) {
		admins, err := client.Admins.List(context.Background(), opts)

		if err != nil {
			t.Fatal(err.Error())
		}

		if len(admins) == 0 {
			t.Fatalf("Expected more than 0 admins")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := client.Admins.Delete(context.Background(), "test", opts)
		if err != nil {
			t.Fatal(err.Error())
		}
	})
}
