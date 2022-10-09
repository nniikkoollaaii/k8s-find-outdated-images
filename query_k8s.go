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

func isFlagSet(flagValue string) bool {
	if flagValue == "" {
		return false
	} else {
		return true
	}
}

func getNamespaces(result *map[string]*NotificationData, ctx *cli.Context, clientset *kubernetes.Clientset) {
	var allNamespaces = getAllNamespaces(clientset)

	filterFlag := ctx.String(filterNamespaceAnnotationFlag.Name) // ToDo: Error Handling not correct value for flag
	var namespaces = filterNamespaces(filterFlag, allNamespaces)

	emailNamespaceAnnotationFlagValue := ctx.String(emailNamespaceAnnotationFlag.Name)
	getNamespaceData(emailNamespaceAnnotationFlagValue, namespaces, result)

}

func getNamespaceData(emailNamespaceAnnotationFlagValue string, namespaces map[string]corev1.Namespace, result *map[string]*NotificationData) {

	for namespaceName, namespaceData := range namespaces {
		notificationData := NotificationData{}

		if isFlagSet(emailNamespaceAnnotationFlagValue) {
			notificationData.Email = namespaceData.Annotations[emailNamespaceAnnotationFlagValue]
		}

		(*result)[namespaceName] = &notificationData
	}
}

func getImages(allImages *map[string]ImageData, namespaces *map[string]*NotificationData, ctx *cli.Context, clientset *kubernetes.Clientset) {

	// iterate over the filtered namespaces
	for namespace := range *namespaces {
		//query all pods in the namespace
		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		log.Debugf("There are %d pods in the namespace %s", len(pods.Items), namespace)

		//iterate over all pods and their images and add to result set
		addImageData(allImages, namespaces, pods)
	}
}

func getAllNamespaces(clientset *kubernetes.Clientset) *corev1.NamespaceList {
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return namespaces
}

func filterNamespaces(filterFlag string, allNamespaces *corev1.NamespaceList) map[string]corev1.Namespace {
	var result = make(map[string]corev1.Namespace)

	if !isFlagSet(filterFlag) {
		for _, namespace := range allNamespaces.Items {
			result[namespace.Name] = namespace
		}
	} else {
		var filterFlagAnnotationKey string
		var filterFlagAnnotationValue string

		if strings.Contains(filterFlag, "=") {
			filterFlagAnnotationKey = strings.Split(filterFlag, "=")[0]
			filterFlagAnnotationValue = strings.Split(filterFlag, "=")[1]
			log.Debugf("Filtering namespaces for annotation with key '%s' and value '%s'", filterFlagAnnotationKey, filterFlagAnnotationValue)
		} else {
			filterFlagAnnotationKey = filterFlag
			log.Debugf("Filtering namespaces for annotation with key '%s'", filterFlagAnnotationKey)
		}

		// iterate over all namespaces and check annotations
		for _, namespace := range allNamespaces.Items {

			if value, ok := namespace.Annotations[filterFlagAnnotationKey]; ok {

				//when filtering only for existence of key
				if filterFlagAnnotationValue == "" {
					// then add the namespace object to the global map
					result[namespace.Name] = namespace
				}

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

func addImageData(allImages *map[string]ImageData, namespaces *map[string]*NotificationData, pods *corev1.PodList) {

	//iterate over all pods and their images and add to result set
	for _, pod := range pods.Items {

		//first the initContainers array
		initContainers := pod.Spec.InitContainers
		for _, initContainer := range initContainers {
			value, exists := (*allImages)[initContainer.Image] //check if image is already in result set
			if !exists {                                       // when not add ImageData struct with initial RunLocation
				(*allImages)[initContainer.Image] = ImageData{
					Image: initContainer.Image,
					Findings: []FindingData{
						{
							Namespace:        pod.Namespace,
							PodName:          pod.Name,
							NotificationData: (*namespaces)[pod.Namespace],
						},
					},
				}
			} else { // when found add only the information where this image is used
				value.Findings = append(value.Findings, FindingData{
					Namespace: pod.Namespace,
					PodName:   pod.Name,
				})
				(*allImages)[initContainer.Image] = value
			}
		}

		// second the containers array
		containers := pod.Spec.Containers
		for _, container := range containers {
			value, exists := (*allImages)[container.Image]
			if !exists {
				(*allImages)[container.Image] = ImageData{
					Image: container.Image,
					Findings: []FindingData{
						{
							Namespace: pod.Namespace,
							PodName:   pod.Name,
						},
					},
				}
			} else {
				value.Findings = append(value.Findings, FindingData{
					Namespace: pod.Namespace,
					PodName:   pod.Name,
				})
				(*allImages)[container.Image] = value

			}
		}

	}
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
