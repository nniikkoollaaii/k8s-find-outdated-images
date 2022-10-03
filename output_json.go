package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/urfave/cli/v2"
)

func outputResult(images *map[string]ImageData, namespaces *map[string]NotificationData, ctx *cli.Context) {

	test := struct {
		Findings   map[string]ImageData
		Namespaces map[string]NotificationData
	}{
		Findings:   *images,
		Namespaces: *namespaces,
	}
	file, _ := json.MarshalIndent(test, "", " ")

	resultFileName := ctx.String(resultFileName.Name)

	_ = ioutil.WriteFile(resultFileName, file, 0644)
}
