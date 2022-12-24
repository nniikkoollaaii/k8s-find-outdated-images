package main

import (
	"encoding/json"
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
	//check that email from pointer to NotificationData is serialized too
	//ToDo: use better method of checking json value
	if !strings.Contains(string(result), "\"Email\": \"test@domain.com\"") {
		t.Fatalf("Output json does not contain expected key")
	}
}

func TestJsonOutputGroupByEmail(t *testing.T) {
	notificationData := NotificationData{}
	notificationData.Email = "test@domain.com"
	notificationData2 := NotificationData{}
	notificationData2.Email = "test2@domain.com"

	images := make(map[string]ImageData)
	images["my.domain.com/image:v1"] = ImageData{
		Image:          "my.domain.com/image:v1",
		BuildTimestamp: time.Now().Add(-(time.Hour * 24 * 30)),
		Findings: []FindingData{
			FindingData{
				Namespace:        "test",
				PodName:          "testpod1",
				NotificationData: &notificationData,
			},
			FindingData{
				Namespace:        "test",
				PodName:          "testpod2",
				NotificationData: &notificationData,
			},
			FindingData{
				Namespace:        "test3",
				PodName:          "testpod3",
				NotificationData: &notificationData,
			},
			FindingData{
				Namespace:        "test",
				PodName:          "testpod4",
				NotificationData: &notificationData2,
			},
		},
	}

	images["my.domain.com/image2:v1"] = ImageData{
		Image:          "my.domain.com/image2:v1",
		BuildTimestamp: time.Now().Add(-(time.Hour * 24 * 30)),
		Findings: []FindingData{
			{
				Namespace:        "test",
				PodName:          "testpod5",
				NotificationData: &notificationData,
			},
			{
				Namespace:        "test",
				PodName:          "testpod6",
				NotificationData: &notificationData2,
			},
		},
	}

	result, err := getJsonGroupedByEmail(&images)
	if err != nil {
		t.Fatalf("Error not nil when serializing to Json")
	}

	var unmarshalledResult ResultGroupedByEmail
	err = json.Unmarshal(result, &unmarshalledResult)
	if err != nil {
		t.Fatalf("Error not nil when deserializing to Json")
	}
	//log.Println(string(result))

	if unmarshalledResult.Notifications["test@domain.com"].Images["my.domain.com/image:v1"].Findings[0].PodName != "testpod1" {
		t.Fatalf("Not found pod1 in group by email output result")
	}

	if unmarshalledResult.Notifications["test@domain.com"].Images["my.domain.com/image:v1"].Findings[1].PodName != "testpod2" {
		t.Fatalf("Not found pod2 in group by email output result")
	}

	if !(unmarshalledResult.Notifications["test@domain.com"].Images["my.domain.com/image:v1"].Findings[2].PodName == "testpod3" &&
		unmarshalledResult.Notifications["test@domain.com"].Images["my.domain.com/image:v1"].Findings[2].Namespace == "test3") {
		t.Fatalf("Not found pod3 in group by email output result")
	}

	if unmarshalledResult.Notifications["test2@domain.com"].Images["my.domain.com/image:v1"].Findings[0].PodName != "testpod4" {
		t.Fatalf("Not found pod4 in group by email output result")
	}

	if unmarshalledResult.Notifications["test@domain.com"].Images["my.domain.com/image2:v1"].Findings[0].PodName != "testpod5" {
		t.Fatalf("Not found pod5 in group by email output result")
	}

	if unmarshalledResult.Notifications["test2@domain.com"].Images["my.domain.com/image2:v1"].Findings[0].PodName != "testpod6" {
		t.Fatalf("Not found pod6 in group by email output result")
	}
}
