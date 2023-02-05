package main

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/urfave/cli/v2"
)

func outputJsonResult(images *map[string]ImageData, ctx *cli.Context) {
	if ctx.String(resultFileFormatFlag.Name) != "json" {
		return
	}

	resultFileName := ctx.String(resultFileNameFlag.Name)

	var fileContent []byte
	var err error
	if ctx.Bool(resultFormatGroupByEmailFlag.Name) {
		fileContent, err = getJsonGroupedByEmail(images)
	} else {
		fileContent, err = getJson(images)
	}

	if err != nil {
		panic(err)
	}

	_ = ioutil.WriteFile(resultFileName, fileContent, 0644)
}

func getJson(images *map[string]ImageData) ([]byte, error) {
	return json.MarshalIndent(images, "", " ")
}
func generateNotificationDataModel(images *map[string]ImageData) ResultGroupedByEmail {
	var result ResultGroupedByEmail
	result.Notifications = make(map[string]ResultGroupedByEmailOutdatedImages)

	for imageName, imageData := range *images {

		for _, finding := range imageData.Findings {

			resultForEmail, existsEntryForEmail := (result.Notifications)[finding.NotificationData.Email]
			if !existsEntryForEmail {
				//add first finding for this notification email address

				images := make(map[string]ResultContentData)
				images[imageName] = ResultContentData{
					BuildTimestamp: getUserStringForBuildTimestamp(imageData.BuildTimestamp),
					Findings: []ResultContentFindingData{
						{
							Namespace: finding.Namespace,
							PodName:   finding.PodName,
						},
					},
				}

				result.Notifications[finding.NotificationData.Email] = ResultGroupedByEmailOutdatedImages{
					Images: images,
				}
			} else {
				//check if the current image already exists in the result
				resultForImage, existsEntryForImage := resultForEmail.Images[imageName]
				if !existsEntryForImage {

					// add new image to result with first finding
					resultForEmail.Images[imageName] = ResultContentData{
						BuildTimestamp: getUserStringForBuildTimestamp(imageData.BuildTimestamp),
						Findings: []ResultContentFindingData{
							{
								Namespace: finding.Namespace,
								PodName:   finding.PodName,
							},
						},
					}
				} else {
					// add finding for image for email
					resultForImage.Findings = append(resultForImage.Findings, ResultContentFindingData{
						Namespace: finding.Namespace,
						PodName:   finding.PodName,
					})
					resultForEmail.Images[imageName] = resultForImage
				}
			}
		}
	}

	return result
}

// convert the image build timestamp to a understandable information for the end user
func getUserStringForBuildTimestamp(timestamp time.Time) string {

	if (timestamp.Equal(time.Time{})) {
		return "UNKOWN"
	}

	return timestamp.Format(time.RFC3339)
}

func getJsonGroupedByEmail(images *map[string]ImageData) ([]byte, error) {
	result := generateNotificationDataModel(images)
	return json.MarshalIndent(result, "", " ")
}
