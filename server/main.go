package main

import (
	"fmt"
	"time"

	"github.com/JaSei/pathutil-go"
	"github.com/alecthomas/kingpin"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type available struct {
	name            string
	ingressStore    cache.Store
	serviceStore    cache.Store
	deploymentStore cache.Store
}

var (
	availableNamespaces map[string]available
	version             = "devel"
	devel               = kingpin.Flag("devel", "development mode (debug log level, CORS allow *)").Bool()
	port                = kingpin.Flag("port", "server port").Default("8080").Uint16()
	kubecfg             = kingpin.Flag("kubeconfig", "kube config file").Default("~/.kube/config").String()
	inCluster           = kingpin.Flag("incluster", "use in cluster config (serviceaccount").Bool()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	var config *rest.Config
	var err error
	if *inCluster {
		// creates the in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		var kubeconfig pathutil.Path
		if *kubecfg == "~/.kube/config" {
			kubeconfig, err = pathutil.Home(".kube", "config")

			if err != nil {
				log.Fatalf("%s not found", kubeconfig)
			}
		} else {
			kubeconfig, err = pathutil.New(*kubecfg)

			if err != nil {
				log.Fatalf("%s not found", *kubecfg)
			}
		}

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig.String())
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	availableNamespaces, err = findAvailableNamespaces(clientset)
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()

	if *devel {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		}))
		e.Logger.SetLevel(log.DEBUG)
	}

	e.Static("/", "static")
	e.GET("/api/namespaces", handleNamespaces)
	e.GET("/api/namespace/:namespace", handleNamespace)
	e.Any("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", *port)))
}

func watchResource(client rest.Interface, namespace, resource string, objType runtime.Object) cache.Store {
	watchlist := cache.NewListWatchFromClient(client, resource, namespace, fields.Everything())

	resyncPeriod := 30 * time.Minute
	//Setup an informer to call functions when the watchlist changes
	eStore, eController := cache.NewInformer(
		watchlist,
		objType,
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{},
	)
	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)
	return eStore
}

func watchPods(clientset *kubernetes.Clientset, namespace string) cache.Store {
	return watchResource(clientset.CoreV1().RESTClient(), namespace, "pods", &v1.Pod{})
}

func watchIngresses(clientset *kubernetes.Clientset, namespace string) cache.Store {
	return watchResource(clientset.Extensions().RESTClient(), namespace, "ingresses", &v1beta1.Ingress{})
}

func watchServices(clientset *kubernetes.Clientset, namespace string) cache.Store {
	return watchResource(clientset.CoreV1().RESTClient(), namespace, "services", &v1.Service{})
}

func watchDeployment(clientset *kubernetes.Clientset, namespace string) cache.Store {
	return watchResource(clientset.Apps().RESTClient(), namespace, "deployments", &appsv1.Deployment{})
}

func findAvailableNamespaces(clientset *kubernetes.Clientset) (availableNamespaces map[string]available, err error) {
	availableNamespaces = make(map[string]available)
	namespaces, err := clientset.Core().Namespaces().List(metav1.ListOptions{})

	if err != nil {
		return availableNamespaces, errors.Wrap(err, "List namespaces failed")
	}

	for _, namespace := range namespaces.Items {
		_, err = clientset.CoreV1().Pods(namespace.GetName()).List(metav1.ListOptions{})

		if k8serrors.IsForbidden(err) {
			log.Warnf("Unable namespace %s", namespace.GetName())
			continue
		} else if err != nil {
			return availableNamespaces, errors.Wrapf(err, "List pods in namespace %s fail", namespace.GetName())
		}

		ingressStore := watchIngresses(clientset, namespace.GetName())
		serviceStore := watchServices(clientset, namespace.GetName())
		deploymentStore := watchDeployment(clientset, namespace.GetName())

		availableNamespaces[namespace.GetName()] = available{namespace.GetName(), ingressStore, serviceStore, deploymentStore}

		log.Infof("Watch namespace %s", namespace.GetName())
	}

	return
}
