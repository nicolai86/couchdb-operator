# couchdb operator

this k8s operator allows you to run a 2.1 couchdb cluster on top of k8s. 

## status

- [ ] operator
    - [ ] CRD (CouchDB)
      - [x] definition
      - [x] management custom object add (spawn cluster)
      - [ ] management custom object update (update cluster)
      - [x] management custom object delete (delete cluster)
    - [x] deployment template (port, readyness, livelyness)
    - [ ] cluster management
      - [ ] new pod -> join cluster
      - [ ] old pod gone -> leave cluster
    - [x] operator definition
- [x] README

## guiding notes

see 7 principles taken from coreOS post: https://coreos.com/blog/introducing-operators.html and https://github.com/kubernetes/community/blob/master/contributors/devel/controllers.md
## prerequisites 

- k8s >= 1.7.0

## usage

```
$ kubectl apply -f k8s/resource-type.yml
$ kubectl apply -f k8s/deployment.yml
```

now, you can deploy a couchdb cluster like this:

```
apiVersion: "stable.couchdb.org/v1"
kind: CouchDB
metadata:
  name: my-couchdb-cluster
spec:
  version: "2.1.0"
  image:   "nicolai86/couchdb"
  replicas: 3 
```

the operator takes care of spawning the necessary pods and joining them into a single cluster.

## development

required: golang 1.9+, kubernetes cluster (minikube, GKE, Azure...).
To get a working operator locally, assuming kubectl is setup correctly:

```
$ go get -u github.com/golang/dep/cmd/dep
$ go get -du github.com/nicolai86/couchdb-operator
$ cd $(go env GOPATH)/src/github.com/nicolai86/couchdb-operator
$ dep ensure
$ OPERATOR_NAMESPACE=couchdb-operator 
$ OPERATOR_NAME=couchdb-operator 
$ KUBECONFIG=~/.kube/config 
$ go run main.go
```