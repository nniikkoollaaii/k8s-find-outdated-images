package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/urfave/cli/v2"
)

func outputJsonResult(images *map[string]ImageData, ctx *cli.Context) {
	resultFormat := ctx.String(resultFileFormat.Name)
	if resultFormat != "json" {
		return
	}

	resultFileName := ctx.String(resultFileName.Name)

	fileContent, err := getJson(images)
	if err != nil {
		panic(err)
	}

	_ = ioutil.WriteFile(resultFileName, fileContent, 0644)
}

func getJson(images *map[string]ImageData) ([]byte, error) {
	return json.MarshalIndent(images, "", " ")
}
