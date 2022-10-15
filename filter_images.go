package main

import "time"

func filterOutdatedImages(allImages *map[string]ImageData, oldestAllowedTimestamp time.Time) {
	for imageName, imageData := range *allImages {
		//if build timestamp is after the oldest allowed timestamp -> than newer and delete from result
		if imageData.BuildTimestamp.After(oldestAllowedTimestamp) {
			delete(*allImages, imageName)
		}
	}
}
