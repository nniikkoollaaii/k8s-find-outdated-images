package main

import (
	"testing"
	"time"
)

func TestFilterImages(t *testing.T) {
	testImages := make(map[string]ImageData)
	testImages["my.domain.com/outdated:1.0.0"] = ImageData{
		Image:          "my.domain.com/outdated:1.0.0",
		BuildTimestamp: time.Now().Add(-(time.Hour * 24 * 40)),
	}
	testImages["my.domain.com/ok:1.0.0"] = ImageData{
		Image:          "my.domain.com/ok:1.0.0",
		BuildTimestamp: time.Now().Add(-(time.Hour * 24 * 20)),
	}
	timestamp := time.Now().Add(-(time.Hour * 24 * 30))

	filterOutdatedImages(&testImages, timestamp)

	if len(testImages) != 1 {
		t.Fatalf("Expected %d outdated image but got %d", 1, len(testImages))
	}
	_, exists := testImages["my.domain.com/outdated:1.0.0"]
	if !exists {
		t.Fatalf("Expected image '%s' is missing", "my.domain.com/outdated:1.0.0")
	}

}
