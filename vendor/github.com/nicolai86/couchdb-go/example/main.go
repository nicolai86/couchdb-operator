package main

//go:generate couchdb-gen -pkg main -file couchdb.go $GOFILE

// Customer represents a real customer within the CRM
// @couchdb @collection
type Customer struct {
}

func main() {

}
