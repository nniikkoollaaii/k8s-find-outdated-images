package main

import (
	"testing"
	"time"
)

func TestEmailAdminTemplating(t *testing.T) {
	notificationData := NotificationData{}
	notificationData.Email = "test@domain.com"

	layout := "2006-01-02T15:04:05.000Z"
	fakeBuildTimestamp, _ := time.Parse(layout, "2022-12-31T13:10:00.000Z")

	testData := ResultGroupedByEmail{
		Notifications: map[string]ResultGroupedByEmailOutdatedImages{
			"test@test.com": ResultGroupedByEmailOutdatedImages{
				Images: map[string]ResultContentData{
					"my.domain.com/image:v1": {
						BuildTimestamp: getUserStringForBuildTimestamp(fakeBuildTimestamp),
						Findings: []ResultContentFindingData{
							{
								Namespace: "test",
								PodName:   "testpod",
							},
							{
								Namespace: "test2",
								PodName:   "testpod2",
							},
						},
					},
					"my.domain.com/image2:v1": {
						BuildTimestamp: getUserStringForBuildTimestamp(fakeBuildTimestamp),
						Findings: []ResultContentFindingData{
							{
								Namespace: "test3",
								PodName:   "testpod3",
							},
						},
					},
				},
			},
		},
	}

	result := templateAdminEmailBodyContent(testData, "", "")
	//log.Println(result.String())

	expectedResult := `
<html>
<head>
<style>
table {
  font-family: arial, sans-serif;
  border-collapse: collapse;
  width: 100%;
}

td, th {
  border: 1px solid #dddddd;
  text-align: left;
  padding: 8px;
}
</style>
</head>
<body>
<p>
The following container images are outdated.
</p>
<p>
</p>
<table>
  <tr>
    <th>Image</th>
    <th>BuildTimestamp</th>
    <th>Namespace</th>
    <th>PodName</th>
  </tr>
  
  
  
  <tr>
    <td>my.domain.com/image2:v1</td>
    <td>2022-12-31T13:10:00Z</td>
    <td>test3</td>
    <td>testpod3</td>
  </tr>
  
  
  
  <tr>
    <td>my.domain.com/image:v1</td>
    <td>2022-12-31T13:10:00Z</td>
    <td>test</td>
    <td>testpod</td>
  </tr>
  
  <tr>
    <td>my.domain.com/image:v1</td>
    <td>2022-12-31T13:10:00Z</td>
    <td>test2</td>
    <td>testpod2</td>
  </tr>
  
  
  
</table>
</body>
</html>
`

	if result.String() != expectedResult {
		t.Fatalf("Notification Email Admin Template Result is wrong.")
	}
}

func TestEmailAdminTemplatingCustomPrefixAndSuffixInContent(t *testing.T) {
	notificationData := NotificationData{}
	notificationData.Email = "test@domain.com"

	layout := "2006-01-02T15:04:05.000Z"
	fakeBuildTimestamp, _ := time.Parse(layout, "2022-12-31T13:10:00.000Z")

	testData := ResultGroupedByEmail{
		Notifications: map[string]ResultGroupedByEmailOutdatedImages{
			"test@test.com": ResultGroupedByEmailOutdatedImages{
				Images: map[string]ResultContentData{
					"my.domain.com/image:v1": {
						BuildTimestamp: getUserStringForBuildTimestamp(fakeBuildTimestamp),
						Findings: []ResultContentFindingData{
							{
								Namespace: "test",
								PodName:   "testpod",
							},
						},
					},
				},
			},
		},
	}

	result := templateAdminEmailBodyContent(testData, "./test/email_content_prefix.tpl", "./test/email_content_suffix.tpl")

	expectedResult := `
<html>
<head>
<style>
table {
  font-family: arial, sans-serif;
  border-collapse: collapse;
  width: 100%;
}

td, th {
  border: 1px solid #dddddd;
  text-align: left;
  padding: 8px;
}
</style>
</head>
<body>
<p>
Test Prefix
</p>
<table>
  <tr>
    <th>Image</th>
    <th>BuildTimestamp</th>
    <th>Namespace</th>
    <th>PodName</th>
  </tr>
  
  
  
  <tr>
    <td>my.domain.com/image:v1</td>
    <td>2022-12-31T13:10:00Z</td>
    <td>test</td>
    <td>testpod</td>
  </tr>
  
  
  
</table>
<p>
Test Suffix
</p>
</body>
</html>
`

	if result.String() != expectedResult {
		t.Fatalf("Notification Email Admin Template Result is wrong.")
	}
}
