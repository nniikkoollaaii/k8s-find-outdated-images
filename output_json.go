package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/urfave/cli/v2"
)

func outputJsonResult(images *map[string]ImageData, ctx *cli.Context) {
	if ctx.String(resultFileFormatFlag.Name) != "json" {
		return
	}

	resultFileName := ctx.String(resultFileNameFlag.Name)

	resultGroupByFlagValue := ctx.String(resultGroupByFlag.Name)
	var fileContent []byte

	if resultGroupByFlagValue == "email" {
		//reorder output struct

	} else {
		// output default struct
		json, err := getJson(images)
		if err != nil {
			panic(err)
		}
		fileContent = json
	}

	_ = ioutil.WriteFile(resultFileName, fileContent, 0644)
}

func getJson(images *map[string]ImageData) ([]byte, error) {
	return json.MarshalIndent(images, "", " ")
}

func getResultPerEmail(images *map[string]ImageData) Result {
	type Result map[string]struct {
		Images map[string]ImageData
	}
	result := make(Result)

	for image := range *images {
		for _, finding := range (*images)[image].Findings {
			emailResult, exists := result[finding.NotificationData.Email]
			if !exists {
				imageData := make(map[string]ImageData)
				imageData[(*images)[image].Image] = ImageData{
					Image:          (*images)[image].Image,
					BuildTimestamp: (*images)[image].BuildTimestamp,
					Findings: []FindingData{
						finding,
					},
				}
			} else {
				currentImageResult := (*images)[image]
				imageResult, exists := emailResult.Images[currentImageResult.Image]
				if !exists {
					imageResult = ImageData{
						Image:          currentImageResult.Image,
						BuildTimestamp: currentImageResult.BuildTimestamp,
						Findings: []FindingData{
							finding,
						},
					}
				} else {
					imageResult.Findings = append(imageResult.Findings, finding)
				}
			}
		}
	}

	return result
}
