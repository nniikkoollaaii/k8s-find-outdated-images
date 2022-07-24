package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"

	// kubernetes client-go
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getK8sClient(ctx *cli.Context) *kubernetes.Clientset {
	contextName := ctx.String(k8sContextFlag.Name)

	kubeConfigPath, err := findKubeConfig()
	if err != nil {
		log.Fatal(err)
	}

	//create a in-mem client config
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath}, //based on the kueconfig at the default path or the ENV
		&clientcmd.ConfigOverrides{
			CurrentContext: contextName, //overridden by the context provided as flag
		}).ClientConfig()
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

func isFilteringEnabled(filterFlag string) bool {
	if filterFlag == "" {
		return false
	} else {
		return true
	}
}

func getImages(allImages *map[string]ImageData, ctx *cli.Context, clientset *kubernetes.Clientset) {

	var namespaces = filterNamespaces(ctx, clientset)

	// iterate over the filtered namespaces
	for _, namespace := range namespaces {
		//query all pods in the namespace
		pods, err := clientset.CoreV1().Pods(namespace.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		log.Debugf("There are %d pods in the namespace %s", len(pods.Items), namespace.Name)

		//iterate over all pods and their images and add to result set
		for _, pod := range pods.Items {

			//first the initContainers array
			initContainers := pod.Spec.InitContainers
			for _, initContainer := range initContainers {
				value, exists := (*allImages)[initContainer.Image] //check if image is already in result set
				if !exists {                                       // when not add ImageData struct with initial RunLocation
					(*allImages)[initContainer.Image] = ImageData{
						Image: initContainer.Image,
						Running: []FindingData{
							FindingData{
								Namespace: pod.Namespace,
								Pod:       pod.Name,
							},
						},
					}
				} else { // when found add only the informationen where this image is used
					value.Running = append(value.Running, FindingData{
						Namespace: pod.Namespace,
						Pod:       pod.Name,
					})
					(*allImages)[initContainer.Image] = value
				}
			}

			// second the containers array
			containers := pod.Spec.Containers
			for _, container := range containers {
				_, ok := (*allImages)[container.Image]
				if !ok {
					(*allImages)[container.Image] = ImageData{
						Image: container.Image,
						Running: []FindingData{
							FindingData{
								Namespace: pod.Namespace,
								Pod:       pod.Name,
							},
						},
					}
				} else {
					value, _ := (*allImages)[container.Image]
					value.Running = append(value.Running, FindingData{
						Namespace: pod.Namespace,
						Pod:       pod.Name,
					})

				}
			}
		}
	}
}

func filterNamespaces(ctx *cli.Context, clientset *kubernetes.Clientset) map[string]corev1.Namespace {
	var result = make(map[string]corev1.Namespace)

	//get all namespaces
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// get the filter flag value
	filterFlag := ctx.String(filterNamespaceAnnotationFlag.Name) // ToDo: Error Handling not correct value for flag

	if !isFilteringEnabled(filterFlag) {
		for _, namespace := range namespaces.Items {
			result[namespace.Name] = namespace
		}
	} else {

		filterFlagAnnotationKey := strings.Split(filterFlag, "=")[0]
		filterFlagAnnotationValue := strings.Split(filterFlag, "=")[1]

		log.Debugf("Filtering namespaces for annotation with key '%s' and value '%s'", filterFlagAnnotationKey, filterFlagAnnotationValue)

		// get labels for each namespace
		for _, namespace := range namespaces.Items {

			if value, ok := namespace.Annotations[filterFlagAnnotationKey]; ok {
				// and the value of the annotation is equal the value specified
				if value == filterFlagAnnotationValue {
					// then add the namespace object to the global map
					result[namespace.Name] = namespace
					log.Debugf("Namespace %s contains annotation with key %s and value %s", namespace.Name, filterFlagAnnotationKey, filterFlagAnnotationValue)
				} else {
					// do nothing
					log.Debugf("Namespace %s contains annotation with key %s but value isn't equals", namespace.Name, filterFlagAnnotationKey)
				}
			} else {
				log.Debugf("Namespace %s doesn't contain a annotation with key %s", namespace.Name, filterFlagAnnotationKey)
			}

		}
	}
	return result
}

// findKubeConfig finds path from env:KUBECONFIG or ~/.kube/config
func findKubeConfig() (string, error) {
	env := os.Getenv("KUBECONFIG")
	if env != "" {
		return env, nil
	}
	path := filepath.Join(homedir.HomeDir(), ".kube", "config")
	return path, nil
}
