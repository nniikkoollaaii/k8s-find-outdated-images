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
	log.Debugf("Found %d namespaces", len(allNamespaces.Items))

	filterFlag := ctx.String(filterNamespaceAnnotationFlag.Name) // ToDo: Error Handling not correct value for flag
	var namespaces = filterNamespaces(filterFlag, allNamespaces)
	log.Debugf("After filtering %d namespaces to check", len(namespaces))

	emailNamespaceAnnotationFlagValue := ctx.String(emailNamespaceAnnotationFlag.Name)
	getNamespaceData(emailNamespaceAnnotationFlagValue, namespaces, result)

}

func getNamespaceData(emailNamespaceAnnotationFlagValue string, namespaces map[string]corev1.Namespace, result *map[string]*NotificationData) {

	for namespaceName, namespaceData := range namespaces {
		notificationData := NotificationData{}

		if isFlagSet(emailNamespaceAnnotationFlagValue) {
			email := namespaceData.Annotations[emailNamespaceAnnotationFlagValue]
			log.Debugf("Notification data email '%s' for namespace '%s'", email, namespaceName)
			notificationData.Email = email
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
		log.Debugf("There are %d pods in the namespace '%s'", len(pods.Items), namespace)

		//iterate over all pods and their images and add to result set
		parsePods(allImages, namespaces, pods)
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
					//and go to next namespace (or next iteration of loop)
					continue
				}

				// and the value of the annotation is equal the value specified
				if value == filterFlagAnnotationValue {
					// then add the namespace object to the global map
					result[namespace.Name] = namespace
					log.Debugf("Namespace '%s' contains annotation with key '%s' and value '%s'", namespace.Name, filterFlagAnnotationKey, filterFlagAnnotationValue)
				} else {
					// do nothing
					log.Debugf("Namespace '%s' contains annotation with key '%s' but value isn't equals", namespace.Name, filterFlagAnnotationKey)
				}
			} else {
				log.Debugf("Namespace '%s' doesn't contain a annotation with key '%s'", namespace.Name, filterFlagAnnotationKey)
			}

		}
	}
	return result
}

func parsePods(allImages *map[string]ImageData, namespaces *map[string]*NotificationData, pods *corev1.PodList) {

	//iterate over all pods and their images and add to result set
	for _, pod := range pods.Items {

		//first the initContainers array
		initContainers := pod.Spec.InitContainers
		for _, initContainer := range initContainers {
			log.Debugf("Found image '%s' in pod '%s' in namespace '%s'", initContainer.Image, pod.Name, pod.Namespace)
			addImageData(initContainer.Image, pod.Name, pod.Namespace, allImages, namespaces)
		}

		// second the containers array
		containers := pod.Spec.Containers
		for _, container := range containers {
			log.Debugf("Found image '%s' in pod '%s' in namespace '%s'", container.Image, pod.Name, pod.Namespace)
			addImageData(container.Image, pod.Name, pod.Namespace, allImages, namespaces)
		}
	}
}

func addImageData(image string, podName string, namespace string, allImages *map[string]ImageData, namespaces *map[string]*NotificationData) {
	value, exists := (*allImages)[image]
	if !exists {
		(*allImages)[image] = ImageData{
			Image: image,
			Findings: []FindingData{
				{
					Namespace:        namespace,
					PodName:          podName,
					NotificationData: (*namespaces)[namespace],
				},
			},
		}
	} else {
		value.Findings = append(value.Findings, FindingData{
			Namespace:        namespace,
			PodName:          podName,
			NotificationData: (*namespaces)[namespace],
		})
		(*allImages)[image] = value
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
