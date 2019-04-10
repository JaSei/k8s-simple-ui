package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
)

func handleNamespaces(c echo.Context) error {
	namespaces := make([]string, len(availableNamespaces))

	i := 0
	for _, n := range availableNamespaces {
		namespaces[i] = n.name
		i++
	}

	return c.JSON(http.StatusOK, namespaces)
}

type ApplicationIngress struct {
	Host          string `json:"host"`
	Path          string `json:"path"`
	LooksLikeGRPC bool   `json:"looks_like_grpc"`
}

type DeploymentStatus struct {
	AvailableReplicas int32 `json:"available_replicas"`
	Replicas          int32 `json:"replicas"`
}

type ApplicationEndpoint struct {
	Name                  string               `json:"name"`
	Ingresses             []ApplicationIngress `json:"ingresses"`
	DeploymentAnnotations map[string]string    `json:"deployment_annotations"`
	DeploymentStatus      DeploymentStatus     `json:"deployment_status"`
	Images                []string             `json:"images"`
}

func handleNamespace(c echo.Context) error {
	requestedNamespace := c.Param("namespace")

	newOut := make([]ApplicationEndpoint, 0)

	namespace, ok := availableNamespaces[requestedNamespace]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Namespace %s not found/available", requestedNamespace))
	}

	err := listWatchedStores(namespace, c.Logger(), func(deployment *appsv1.Deployment, appIngress []ApplicationIngress) {
		deploymentAnnotations := deployment.Spec.Template.GetAnnotations()

		images := make([]string, len(deployment.Spec.Template.Spec.Containers))
		for i, c := range deployment.Spec.Template.Spec.Containers {
			images[i] = c.Image
		}

		newOut = append(newOut, ApplicationEndpoint{
			Name:                  deployment.Name,
			Ingresses:             appIngress,
			DeploymentAnnotations: deploymentAnnotations,
			DeploymentStatus: DeploymentStatus{
				Replicas:          deployment.Status.Replicas,
				AvailableReplicas: deployment.Status.AvailableReplicas,
			},
			Images: images,
		})
	})

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, newOut)
}

func listWatchedStores(namespace available, logger echo.Logger, fnAppend func(*appsv1.Deployment, []ApplicationIngress)) error {
	for _, deploymentObj := range namespace.deploymentStore.List() {
		deployment, ok := deploymentObj.(*appsv1.Deployment)
		if !ok {
			return errors.Errorf("Is not possible type %+v to appsv1.Deployment", deployment)
		}

		services, err := findServicesRelatedTo(namespace, deployment)
		if err != nil {
			return errors.Wrap(err, "findServicesRelatedTo failed")
		}

		appIng := make([]ApplicationIngress, 0)

		if len(services) == 0 {
			logger.Debugf("Deployment %s/%s doesn't have service", deployment.Namespace, deployment.Name)
			fnAppend(deployment, appIng)
			continue
		}

		for _, service := range services {
			ingresses, err := findIngressesRelatedTo(namespace, service)
			if err != nil {
				return errors.Wrap(err, "findIngressesRelatedTo failed")
			}

			if len(ingresses) == 0 {
				logger.Debugf("[%s] Deployment %s, Service %s doesn't have ingress", deployment.Namespace, deployment.Name, service.Name)

				fnAppend(deployment, appIng)
				continue
			}

			for _, ingress := range ingresses {
				for _, rule := range ingress.Spec.Rules {
					for _, path := range rule.HTTP.Paths {
						appIng = append(appIng, ApplicationIngress{
							Host:          rule.Host,
							Path:          path.Path,
							LooksLikeGRPC: looksLikeGRPCIngress(ingress),
						})
					}
				}
			}

			fnAppend(deployment, appIng)
		}
	}

	return nil
}

func looksLikeGRPCIngress(ingress *v1beta1.Ingress) bool {
	for k, v := range ingress.Annotations {
		if strings.HasSuffix(k, "grpc-backend") {
			if v == "true" {
				return true
			}
			break
		}
	}

	return false
}

func findServicesRelatedTo(a available, deployment *appsv1.Deployment) (services []*v1.Service, err error) {
	for _, serviceObj := range a.serviceStore.List() {
		service := serviceObj.(*v1.Service)

		if labels.Equals(service.Spec.Selector, deployment.Spec.Selector.MatchLabels) {
			services = append(services, service)
		}
	}

	return
}

func findIngressesRelatedTo(a available, service *v1.Service) (ingresses []*v1beta1.Ingress, err error) {
	for _, ingressObj := range a.ingressStore.List() {
		ingress, ok := ingressObj.(*v1beta1.Ingress)
		if !ok {
			err = errors.Errorf("Is not possible type %+v to v1beta1.Ingress", ingress)
			return
		}

	RulesPathsLoop:
		for _, rule := range ingress.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				if path.Backend.ServiceName == service.Name {
					ingresses = append(ingresses, ingress)
					break RulesPathsLoop
				}
			}
		}
	}

	return
}
