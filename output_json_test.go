package main

import (
	"strings"
	"testing"
	"time"
)

func TestJsonOutput(t *testing.T) {
	notificationData := NotificationData{}
	notificationData.Email = "test@domain.com"

	images := make(map[string]ImageData)
	images["my.domain.com/image:v1"] = ImageData{
		Image:          "my.domain.com/image:v1",
		BuildTimestamp: time.Now().Add(-(time.Hour * 24 * 30)),
		Findings: []FindingData{
			FindingData{
				Namespace:        "test",
				PodName:          "testpod",
				NotificationData: &notificationData,
			},
		},
	}

	result, err := getJson(&images)

	if err != nil {
		t.Fatalf("Error not nil when serializing to Json")
	}
	t.Log(string(result))
	//check that email from pointer to NotificationData is serialized too
	//ToDo: use better method of checking json value
	if !strings.Contains(string(result), "\"Email\": \"test@domain.com\"") {
		t.Fatalf("Output json does not contain expected key")
	}
}
