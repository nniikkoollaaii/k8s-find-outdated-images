package main

import (
	"os"
	"time"

	str2duration "github.com/xhit/go-str2duration/v2"

	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
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

var resultFileName = cli.StringFlag{
	Name:        "output",
	Aliases:     []string{"o"},
	DefaultText: "result.json",
	Usage:       "The name of the file to write the result",
	Required:    false,
	Value:       "result.json",
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
					&resultFileName,
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
	Findings       []FindingData
	Image          string
	BuildTimestamp time.Time
}

// information where the image has been found
type FindingData struct {
	Namespace        string
	PodName          string
	NotificationData *NotificationData
}

type NotificationData struct {
	Email string // email address to notify, when image is outdated
	//ToDo: Support more notification methods
}

/**
* Template method
 */
func findOutdatedImages(ctx *cli.Context) error {

	images := make(map[string]ImageData)
	namespaces := make(map[string]*NotificationData)

	// Preprare:
	var k8sclient = getK8sClient(ctx)
	allowedAge, err := str2duration.ParseDuration(ctx.String(ageFlag.Name))
	if err != nil {
		log.Errorf("Cannot parse allowed age from \"%s\" for flag \"--%s\"", ctx.String(ageFlag.Name), ageFlag.Name)
	}
	oldestAllowedTimestamp := time.Now().Add(-allowedAge)

	//0. Step: Get all namespaces with relevant data and filter them
	getNamespaces(&namespaces, ctx, k8sclient)

	//1. Step: Get all (filtered) container images running in the cluster
	getImages(&images, &namespaces, ctx, k8sclient)

	//2. Step: Query Registry for Build-Timestamp of each image
	queryTimestamps(&images)

	log.Info(&images)

	//3. Step: filter Images which are outdated
	filterOutdatedImages(&images, oldestAllowedTimestamp)

	//4. Step: Output results
	outputJsonResult(&images, ctx)
	return nil
}
