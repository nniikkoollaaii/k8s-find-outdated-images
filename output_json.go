package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/urfave/cli/v2"
)

func outputResult(images *map[string]ImageData, ctx *cli.Context) {
	file, _ := json.MarshalIndent(*images, "", " ")

	resultFileName := ctx.String(resultFileName.Name)

	_ = ioutil.WriteFile(resultFileName, file, 0644)
}
