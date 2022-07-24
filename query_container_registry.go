package main

import (
	"time"

	// go-containerregistry
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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
