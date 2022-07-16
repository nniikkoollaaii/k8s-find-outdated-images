package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"

	//"encoding/csv"
	"os"
	"path/filepath"
	"time"

	str2duration "github.com/xhit/go-str2duration/v2"

	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"

	// kubernetes client-go
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	// go-containerregistry
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

var verboseFlag = cli.BoolFlag{
	Name:    "verbose",
	Aliases: []string{"v"},
	Usage:   "enables debugging log level",
}
var versionFlag = cli.BoolFlag{
	Name:    "print-version",
	Aliases: []string{"V"},
	Usage:   "print only the version",
}

var k8sContextFlag = cli.StringFlag{
	Name:        "context",
	Aliases:     []string{"c"},
	DefaultText: "aks-test",
	Usage:       "Context name in your KUBECONFIG",
	Required:    true,
	Value:       "aks-test",
}

var ageFlag = cli.StringFlag{
	Name:        "age",
	Aliases:     []string{"a"},
	DefaultText: "30d",
	Usage:       "The max allowed age of an image",
	Required:    true,
	Value:       "30d",
}

var filterNamespaceAnnotationFlag = cli.StringFlag{
	Name:        "filter",
	Aliases:     []string{"f"},
	DefaultText: "type=workload",
	Usage:       "Filter on namespaces containing the annotation and value. Without this filter all namespaces are checked",
	Required:    false,
	Value:       "",
}

var emailNamespaceAnnotationFlag = cli.StringFlag{
	Name: "email",
	//Aliases:     []string{"e"},
	DefaultText: "email",
	Usage:       "The annotation key on the namespaces containing an email address to contact if there are outdated images used in this namespace",
	Required:    false,
	Value:       "",
}

func main() {

	app := &cli.App{
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&verboseFlag,
			&versionFlag,
		},
		Before: func(ctx *cli.Context) error {
			verboseFlagValue := ctx.Bool(verboseFlag.Name)
			if verboseFlagValue {
				log.SetLevel(log.DebugLevel)
			}
			return nil
		},
		Commands: []*cli.Command{
			&cli.Command{
				Name:      "find",
				Usage:     "do it!",
				UsageText: "find - does the finding",
				//Description: "no really, there is a lot of dooing to be done",
				//ArgsUsage:   "[arrgh]",
				Flags: []cli.Flag{
					&k8sContextFlag,
					&ageFlag,
					&filterNamespaceAnnotationFlag,
					&emailNamespaceAnnotationFlag,
				},
				Action: func(c *cli.Context) error {
					return findOutdatedImages(c)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// DTO containing the data collected by this tool
type ImageData struct {
	Running        []FindingData
	Image          string
	BuildTimestamp time.Time
}

// information where the image has been found
type FindingData struct {
	Namespace string
	Pod       string
	Email     string // email address to notify, when image is outdated
}

var Namespaces = make(map[string]corev1.Namespace)
var AllImages = make(map[string]ImageData)

/**
* Template method
 */
func findOutdatedImages(ctx *cli.Context) error {
	// Preprare:
	var k8sclient = getK8sClient(ctx)
	allowedAge, err := str2duration.ParseDuration(ctx.String(ageFlag.Name))
	if err != nil {
		log.Errorf("Cannot parse allowed age from \"%s\" for flag \"--%s\"", ctx.String(ageFlag.Name), ageFlag.Name)
	}
	oldestAllowedTimestamp := time.Now().Add(-allowedAge)

	//1. Step: Get all container images running in the cluster
	AllImages = getAllImages(ctx, k8sclient)

	//2. Step: Filter / ...

	//3. Step: Query Registry for Build-Timestamp of the image
	for key, value := range AllImages {
		value.BuildTimestamp = getContainerCreatedTimestampForImage(key)
		AllImages[key] = value
	}

	//4. Step: Output results
	log.Info(AllImages)
	outputImagesOlderThanAllowed(oldestAllowedTimestamp)
	//jsonString, _ := json.MarshalIndent(AllImages, "", "    ")
	//log.Infoln(string(jsonString))
	return nil
}

func outputImagesOlderThanAllowed(oldestAllowedTimestamp time.Time) {

	file, _ := json.MarshalIndent(AllImages, "", " ")

	_ = ioutil.WriteFile("result.json", file, 0644)
}

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

func getAllImages(ctx *cli.Context, clientset *kubernetes.Clientset) map[string]ImageData {
	//get all namespaces
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	// get the filter flag value
	filterFlag := ctx.String(filterNamespaceAnnotationFlag.Name) // ToDo: Error Handling not correct value for flag
	isFilteringEnabled := true
	var filterFlagAnnotationKey, filterFlagAnnotationValue string
	if filterFlag == "" {
		isFilteringEnabled = false
	} else {
		filterFlagAnnotationKey = strings.Split(filterFlag, "=")[0]
		filterFlagAnnotationValue = strings.Split(filterFlag, "=")[1]
	}

	log.Debugf("Filtering namespaces for annotation with key '%s' and value '%s'", filterFlagAnnotationKey, filterFlagAnnotationValue)

	// get labels for each namespace
	for _, namespace := range namespaces.Items {

		if !isFilteringEnabled {

			Namespaces[namespace.Name] = namespace

		} else {

			if value, ok := namespace.Annotations[filterFlagAnnotationKey]; ok {
				// and the value of the annotation is equal the value specified
				if value == filterFlagAnnotationValue {
					// then add the namespace object to the global map
					Namespaces[namespace.Name] = namespace
					log.Debugf("Namespace %s contains annotation with key %s and value %s", namespace.Name, filterFlagAnnotationKey, filterFlagAnnotationValue)
				} else {
					// do nothing
					log.Debugf("Namespace %s contains annotation with key %s but value isn't equal", namespace.Name)
				}
			} else {
				log.Debugf("Namespace %s doesn't contain a annotation with key %s", namespace.Name, filterFlagAnnotationKey)
			}
		}

	}

	resultSet := make(map[string]ImageData)

	// now iterate over the filtered namespaces
	for _, namespace := range Namespaces {
		//query all pods in the namespaces
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
				value, exists := resultSet[initContainer.Image] //check if image is already in result set
				if !exists {                                    // when not add ImageData struct with initial RunLocation
					resultSet[initContainer.Image] = ImageData{
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
					resultSet[initContainer.Image] = value
				}
			}

			// next the containers array
			containers := pod.Spec.Containers
			for _, container := range containers {
				_, ok := resultSet[container.Image]
				if !ok {
					resultSet[container.Image] = ImageData{
						Image: container.Image,
						Running: []FindingData{
							FindingData{
								Namespace: pod.Namespace,
								Pod:       pod.Name,
							},
						},
					}
				} else {
					value, _ := resultSet[container.Image]
					value.Running = append(value.Running, FindingData{
						Namespace: pod.Namespace,
						Pod:       pod.Name,
					})

				}
			}
		}
	}

	return resultSet
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

func getContainerCreatedTimestampForImage(image string) time.Time {
	ref, err := name.ParseReference(image)
	if err != nil {
		panic(err)
	}

	// Fetch the manifest using default credentials.
	// "The DefaultKeychain will use credentials as described in your Docker config file -- usually ~/.docker/config.json, or %USERPROFILE%\.docker\config.json on Windows -- or the location described by the DOCKER_CONFIG environment variable, if set."
	// https://github.com/google/go-containerregistry/blob/main/pkg/authn/README.md
	descriptor, err := remote.Get(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		panic(err)
	}

	img, err := descriptor.Image() // using amd64/linux as default platform
	if err != nil {
		panic(err)
	}
	configFile, err := img.ConfigFile()
	if err != nil {
		panic(err)
	}

	return configFile.Created.Time

	//log.Info("Image: " + image + "\n")
	//log.Info(img.Manifest)

	// Prints the digest of registry.example.com/private/repo
	//fmt.Println(img.Digest)
}
