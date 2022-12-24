package main

import (
	"time"
)

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

// Data annotated on the namespace level to contact the image owner
type NotificationData struct {
	Email string // email address to notify, when image is outdated
	//ToDo: Support more notification methods
}

// customized result output format
type ResultGroupedByEmail struct {
	//key: email address
	Notifications map[string]ResultGroupedByEmailOutdatedImages
}
type ResultGroupedByEmailOutdatedImages struct {
	//key: image name
	Images map[string]ResultContentData
}
type ResultContentData struct {
	BuildTimestamp time.Time
	Findings       []ResultContentFindingData
}

type ResultContentFindingData struct {
	Namespace string
	PodName   string
}
