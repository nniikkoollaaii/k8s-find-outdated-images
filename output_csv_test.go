package main

import (
	"bytes"
	"encoding/csv"
	"testing"
	"time"
)

func TestCsvOutput(t *testing.T) {
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

	result, err := getCsv(&images)

	if err != nil {
		t.Fatalf("Error not nil when serializing to Json")
	}
	reader := bytes.NewReader(result)
	csvReader := csv.NewReader(reader)

	records, err := csvReader.ReadAll()

	if err != nil {
		t.Fatalf("Error not nil when serializing to Csv")
	}
	if records[1][3] != "test@domain.com" {
		t.Fatalf("Output csv does not contain expected entry")
	}
}
