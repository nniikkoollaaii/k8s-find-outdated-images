package main

import "time"

func filterOutdatedImages(allImages *map[string]ImageData, oldestAllowedTimestamp time.Time) {
	for imageName, imageData := range *allImages {
		if imageData.BuildTimestamp.Before(oldestAllowedTimestamp) {
			delete(*allImages, imageName)
		}
	}
}
