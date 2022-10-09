package main

import (
	"bytes"
	"encoding/csv"
	"io/ioutil"
	"time"

	"github.com/urfave/cli/v2"
)

func outputCsvResult(images *map[string]ImageData, ctx *cli.Context) {
	if ctx.String(resultFileFormatFlag.Name) != "csv" {
		return
	}

	resultFileName := ctx.String(resultFileNameFlag.Name)

	fileContent, err := getCsv(images)
	if err != nil {
		panic(err)
	}

	_ = ioutil.WriteFile(resultFileName, fileContent, 0644)
}

func getCsv(images *map[string]ImageData) ([]byte, error) {
	b := new(bytes.Buffer)
	w := csv.NewWriter(b)

	//ToDO
	csvObject := [][]string{}
	csvObject = append(csvObject, []string{
		"Image",
		"Namespace",
		"BuildTimestamp",
		"Email",
	})
	for image := range *images {
		for _, finding := range (*images)[image].Findings {
			csvObject = append(csvObject, []string{
				image,             //Image
				finding.Namespace, //Namespace
				(*images)[image].BuildTimestamp.Format(time.RFC3339), //BuildTimestamp
				finding.NotificationData.Email,                       //Email
			})
		}
	}
	w.WriteAll(csvObject)

	return b.Bytes(), w.Error()
}
