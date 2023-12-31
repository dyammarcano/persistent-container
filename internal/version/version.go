// Code generated by the go generate ./... command; DO NOT EDIT.

package version

import "fmt"

var ver *Version

func init() {
	ver = &Version{
		Version:    "v0.0.1-dev",
		CommitHash: "cfd227ab702a567ca15e12388e98d4ff7e9e99b5",
		Date:       "2023-12-04T12:30:59Z",
		Signature:  "EiJVk93cW54E7UrDipoQP6ffhrSTwHwRNds9brHis4z",
	}
}

type Version struct {
	Version    string `json:"version"`
	CommitHash string `json:"commitHash"`
	Date       string `json:"date"`
	Signature  string `json:"signature"`
}

func GetVersion() *Version {
	return ver
}

func GetVersionString() string {
	return fmt.Sprintf("version: %s, commit: %s, date: %s, uuid: %s", ver.Version, ver.CommitHash, ver.Date, ver.Signature)
}
