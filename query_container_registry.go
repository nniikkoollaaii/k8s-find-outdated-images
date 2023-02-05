package main

import (
	"time"

	// go-containerregistry
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	log "github.com/sirupsen/logrus"
)

func queryTimestamps(allImages *map[string]ImageData) {
	for key, value := range *allImages {
		value.BuildTimestamp = getImageCreatedTimestampForImage(key)
		(*allImages)[key] = value
	}
}

func getImageCreatedTimestampForImage(image string) time.Time {
	ref, err := name.ParseReference(image)
	if err != nil {
		panic(err)
	}

	// Fetch the manifest using default credentials.
	// "The DefaultKeychain will use credentials as described in your Docker config file -- usually ~/.docker/config.json, or %USERPROFILE%\.docker\config.json on Windows -- or the location described by the DOCKER_CONFIG environment variable, if set."
	// https://github.com/google/go-containerregistry/blob/main/pkg/authn/README.md
	descriptor, err := remote.Get(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {

		//ToDo: How to differentiate between different failure states?

		//When the error is returned because the Image ref does not exist -> only log a warning
		log.Warn(err)
		//Case 1: The pod is in Pending state because of wrong image reference (image does simply not exist because not built or pushed yet)
		//Case 2: The image in the pod is so old it does not exist in the registry anymore (because of housekeeping or something like this)
		// -> so assume this Image is outdated
		return time.Time{} // 0001-01-01 00:00:00 +0000 UTC

		//When the error is returned because of a AuthN problem -> the tool should exit ...
		//panic(err)
		//Room for improvement here
	}

	img, err := descriptor.Image() // using amd64/linux as default platform
	if err != nil {
		panic(err)
	}
	configFile, err := img.ConfigFile()
	if err != nil {
		panic(err)
	}

	log.Debugf("Build timestamp for image '%s' is '%s'", image, configFile.Created.Time)

	return configFile.Created.Time

	//log.Info("Image: " + image + "\n")
	//log.Info(img.Manifest)

	// Prints the digest of registry.example.com/private/repo
	//fmt.Println(img.Digest)
}
