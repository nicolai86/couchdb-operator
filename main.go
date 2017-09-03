package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"

	couchdb "github.com/nicolai86/couchdb-go"
	"github.com/nicolai86/couchdb-operator/probe"
	"github.com/nicolai86/couchdb-operator/spec"
	"github.com/nicolai86/couchdb-operator/version"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	couchdbVersion = "2.1.0"
	couchdbImage   = "nicolai86/couchdb"
	namespace      string
	podName        string
	listenAddr     string
)

func crdRestClient(config rest.Config) (*rest.RESTClient, error) {
	config.GroupVersion = &schema.GroupVersion{"stable.couchdb.org", "v1"}
	scheme := k8sruntime.NewScheme()
	scheme.AddKnownTypes(*config.GroupVersion,
		&spec.CouchDB{},
		&spec.CouchDBList{},
	)
	config.APIPath = "/apis"
	config.ContentType = k8sruntime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}
	return rest.RESTClientFor(&config)
}

func main() {
	namespace = os.Getenv("OPERATOR_NAMESPACE")
	if len(namespace) == 0 {
		log.Fatalf("Missing OPERATOR_NAMESPACE")
	}
	podName = os.Getenv("OPERATOR_NAME")
	if len(podName) == 0 {
		log.Fatalf("Missing OPERATOR_NAME")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	go func() {
		log.Printf("received signal: %v", <-c)
		os.Exit(1)
	}()

	log.Printf("couchdb-operator Version: %v", version.Version)
	log.Printf("Git SHA: %s", version.GitSHA)
	log.Printf("Go Version: %s", runtime.Version())
	log.Printf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)

	kubeconfig := ""
	flag.StringVar(&listenAddr, "listen-addr", "0.0.0.0:8080", "The address on which the HTTP server will listen to")
	flag.StringVar(&kubeconfig, "kubeconfig", kubeconfig, "kubeconfig file")
	flag.Parse()

	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	var (
		config *rest.Config
		err    error
	)

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating client: %v", err)
		os.Exit(1)
	}

	http.HandleFunc(probe.HTTPReadyzEndpoint, probe.ReadyzHandler)
	go http.ListenAndServe(listenAddr, nil)

	client := kubernetes.NewForConfigOrDie(config)

	couchRestClient, err := crdRestClient(*config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating client: %v", err)
		os.Exit(1)
	}
	{
		source := cache.NewListWatchFromClient(
			client.Core().RESTClient(),
			apiv1.ResourcePods.String(),
			apiv1.NamespaceAll,
			fields.Everything())

		_, controller := cache.NewInformer(
			source,
			&apiv1.Pod{},
			0,

			cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					c, ok := obj.(*apiv1.Pod)
					if !ok {
						return
					}
					if c.Labels["app"] != "couchdb" {
						return
					}
					log.Printf("pod %#v creation in cluster %q\n", c.UID, c.Labels["cluster"])
					// TODO check if too many pods. if so, delete
				},
				UpdateFunc: func(old interface{}, new interface{}) {
					c, ok := new.(*apiv1.Pod)
					if !ok {
						return
					}
					if c.Labels["app"] != "couchdb" {
						return
					}
					log.Printf("pod %#v (%s) update in cluster %q\n", c.UID, c.Status.Phase, c.Labels["cluster"])

					res := couchRestClient.Get().Namespace(c.Namespace).Resource("couchdbs").Name(c.Labels["cluster"]).Do()
					var cluster *spec.CouchDB
					if o, err := res.Get(); err != nil {
						// log.Printf("failed to lookup couchdb: %v", err.Error())
						return
					} else {
						cluster = o.(*spec.CouchDB)
					}

					list, err := client.CoreV1().Pods(c.Namespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("cluster=%s", c.Labels["cluster"])})
					if err != nil {
						log.Printf("could nod list couchdb cluster %q pods: %v\n", c.Name, err.Error())
						return
					}
					if cluster.Annotations["couchdb.org/initialized"] == "true" {
						// TODO check if node needs to be added
						return
					}
					if len(list.Items) != cluster.Spec.Size {
						return
					}
					ready := true
					for _, p := range list.Items {
						ready = ready && p.Status.Phase == apiv1.PodRunning
						for _, status := range p.Status.ContainerStatuses {
							ready = ready && status.State.Running != nil
							ready = ready && status.Ready
							log.Printf("pod %q status %t, %#v\n", p.UID, status.Ready, status.State.Running)
						}
					}
					if !ready {
						log.Printf("Not ready to init cluster")
						return
					}
					log.Printf("initializing cluster...")
					// mark as initialized
					defer func() {
						annotations := cluster.Annotations
						annotations["couchdb.org/initialized"] = "true"
						cluster.SetAnnotations(annotations)
						res := couchRestClient.Put().Namespace(cluster.Namespace).Resource("couchdbs").Name(cluster.Name).Body(cluster).Do()
						log.Printf("writing resource annotations")
						if err := res.Error(); err != nil {
							log.Printf("failed to update cluster state: %#v", err.Error())
						}
					}()

					// TODO ensure clustering is already enabled!
					// for _, p := range list.Items[1:] {

					// 	c, _ := couchdb.New(fmt.Sprintf("http://%s:5984", p.Status.PodIP), &http.Client{}, couchdb.WithBasicAuthentication("admin", "admin"))
					// 	if err := c.Cluster.BeginSetup(couchdb.SetupOptions{
					// 		BindAddress: "0.0.0.0",
					// 		Username:    "admin",
					// 		Password:    "admin",
					// 		NodeCount:   len(list.Items),
					// 	}); err != nil {
					// 		log.Printf("failed to start cluster setup: %v\n", err.Error())
					// 	}
					// }

					{
						setup := list.Items[0]
						log.Printf("ready to initialize cluster %q/w", cluster.Name, setup.Status.PodIP)
						c, _ := couchdb.New(fmt.Sprintf("http://%s:5984", setup.Status.PodIP), &http.Client{}, couchdb.WithBasicAuthentication("admin", "admin"))
						for _, p := range list.Items[1:] {
							// if err := c.Cluster.BeginSetup(couchdb.SetupOptions{
							// 	BindAddress:    "0.0.0.0",
							// 	Username:       "admin",
							// 	Password:       "admin",
							// 	NodeCount:      len(list.Items),
							// 	Port:           15984,
							// 	RemoteNode:     p.Status.PodIP,
							// 	RemotePassword: "admin",
							// 	RemoteUsername: "admin",
							// }); err != nil {
							// 	log.Printf("begin setup for node %s failed: %v\n", p.Status.PodIP, err.Error())
							// }
							if err := c.Cluster.AddNode(couchdb.AddNodeOptions{
								Host:     p.Status.PodIP,
								Username: "admin",
								Password: "admin",
								Port:     5984,
							}); err != nil {
								log.Printf("add node for node %s failed: %v\n", p.Status.PodIP, err.Error())
							}
						}
					}

					// for _, p := range list.Items {
					// 	c, _ := couchdb.New(fmt.Sprintf("http://%s:5984", p.Status.PodIP), &http.Client{}, couchdb.WithBasicAuthentication("admin", "admin"))
					// 	if err := c.Cluster.EndSetup(); err != nil {
					// 		log.Printf("failed to finish cluster setup: %v\n", err.Error())
					// 	}
					// }
				},
				DeleteFunc: func(obj interface{}) {
					c, ok := obj.(*apiv1.Pod)
					if !ok {
						return
					}
					if c.Labels["app"] != "couchdb" {
						return
					}
					log.Printf("pod %#v deletion in cluster %q\n", c.UID, c.Labels["cluster"])
					// TODO check if cluster exists & needs more. if so, spawn
					// TODO check if cluster exists. if so, remove node
				},
			})
		go controller.Run(nil)
	}
	// TODO new controller watching for couchdb server pods
	{
		source := cache.NewListWatchFromClient(
			couchRestClient,
			"couchdbs",
			apiv1.NamespaceAll,
			fields.Everything())

		_, controller := cache.NewInformer(
			source,
			&spec.CouchDB{},

			// resyncPeriod
			// Every resyncPeriod, all resources in the cache will retrigger events.
			// Set to 0 to disable the resync.
			0,

			cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					c, ok := obj.(*spec.CouchDB)
					if !ok {
						return
					}
					log.Printf("Adding couchdb cluster %q in ns %q\n", c.Name, c.Namespace)
					annotations := c.Annotations
					annotations["couchdb.org/initialized"] = "false"
					c.SetAnnotations(annotations)
					res := couchRestClient.Put().Namespace(c.Namespace).Resource("couchdbs").Name(c.Name).Body(c).Do()
					log.Printf("writing resource annotations")
					if err := res.Error(); err != nil {
						log.Printf("failed to update cluster state: %#v", err.Error())
					}

					list, err := client.CoreV1().Pods(c.Namespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("cluster=%s", c.Name)})
					if err != nil {
						log.Printf("could nod list couchdb cluster %q pods: %v\n", c.Name, err.Error())
						return
					}
					log.Printf("got %d pods in cluster %q in ns %q\n", len(list.Items), c.Name, c.Namespace)

					for i := 0; i < c.Spec.Size-len(list.Items); i++ {
						log.Printf("creating pod %d for cluster %q in ns %q\n", i, c.Name, c.Namespace)
						pod := newCouchdbPod(c.Name, "admin", c.Spec.Pod)
						_, err = client.CoreV1().Pods(c.Namespace).Create(pod)
						if err != nil {
							log.Printf("failed to start pod: %#v", err.Error())
						}
					}
				},
				UpdateFunc: func(old interface{}, new interface{}) {
					oC, ok := old.(*spec.CouchDB)
					if !ok {
						return
					}
					nC, ok := new.(*spec.CouchDB)
					if !ok {
						return
					}
					log.Printf("on Update: %#v -> %#v\n", oC.UID, nC.UID)
					// TODO list deployments matching cluster selector
					// TODO update deployment with new configuration if exists
				},
				DeleteFunc: func(obj interface{}) {
					c, ok := obj.(*spec.CouchDB)
					if !ok {
						return
					}
					log.Printf("Removing couchdb cluster %q in ns %q\n", c.Name, c.Namespace)

					list, err := client.CoreV1().Pods(c.Namespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("cluster=%s", c.Name)})
					if err != nil {
						log.Printf("could not delete clust %q: %v\n", c.ClusterName, err.Error())
					}

					log.Printf("got %d pods in cluster %q in ns %q\n", len(list.Items), c.Name, c.Namespace)
					for i, p := range list.Items {
						log.Printf("deleting pod %d for cluster %q in ns %q\n", i, c.Name, c.Namespace)
						err := client.CoreV1().Pods(c.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
						if err != nil {
							log.Printf("could not delete pod %q: %v\n", p.UID, err.Error())
						}
					}
					// TODO list deployments matching cluster selector
					// TODO if deployments exist, delete all
				},
			})
		controller.Run(nil)
	}

	probe.SetReady()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
}

func couchdbContainer(baseImage, version string) apiv1.Container {
	c := apiv1.Container{
		Name:  "couchdb",
		Image: fmt.Sprintf("%s:%s", baseImage, version),
		Env: []apiv1.EnvVar{
			{
				Name:  "COUCHDB_USER",
				Value: "admin",
			},
			{
				Name:  "COUCHDB_PASSWORD",
				Value: "admin",
			},
			{
				Name: "NODENAME",
				ValueFrom: &apiv1.EnvVarSource{
					FieldRef: &apiv1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
		},
		Ports: []apiv1.ContainerPort{
			{
				Name:          "node-local",
				ContainerPort: int32(5986),
				Protocol:      apiv1.ProtocolTCP,
			},
			{
				Name:          "standalone",
				ContainerPort: int32(5984),
				Protocol:      apiv1.ProtocolTCP,
			},
			{
				Name:          "epmd",
				ContainerPort: int32(4369),
				Protocol:      apiv1.ProtocolTCP,
			},
			{
				Name:          "inet",
				ContainerPort: int32(9100),
				Protocol:      apiv1.ProtocolTCP,
			},
		},
		LivenessProbe: &apiv1.Probe{
			InitialDelaySeconds: 20,
			Handler: apiv1.Handler{
				Exec: &apiv1.ExecAction{
					Command: []string{"pidof", "beam.smp"},
				},
			},
		},
		ReadinessProbe: &apiv1.Probe{
			InitialDelaySeconds: 20,
			Handler: apiv1.Handler{
				HTTPGet: &apiv1.HTTPGetAction{
					Scheme: "HTTP",
					Port:   intstr.FromString("standalone"),
				},
			},
		},
	}

	return c
}

func getMyPodServiceAccount(kubecli kubernetes.Interface) (string, error) {
	var sa string
	pod, err := kubecli.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	sa = pod.Spec.ServiceAccountName

	return sa, err
}

func newCouchdbPod(clustername, password string, spec *spec.PodPolicy) *apiv1.Pod {
	c := couchdbContainer(couchdbImage, couchdbVersion)
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "couchdb-",
			Labels: map[string]string{
				"app":     "couchdb",
				"cluster": clustername,
			},
			Annotations: map[string]string{},
		},
		Spec: apiv1.PodSpec{
			RestartPolicy: apiv1.RestartPolicyAlways,
			Containers:    []apiv1.Container{c},
			DNSPolicy:     apiv1.DNSClusterFirstWithHostNet,
			Subdomain:     clustername,
			Volumes:       nil,
		},
	}
	return pod
}
