// Code generated by the go generate ./... command; DO NOT EDIT.

package version

import "fmt"

var ver *Version

func init() {
	ver = &Version{
		Version:    "v0.0.1-dev",
		CommitHash: "6eb437e1049a7f1fd60f0177c3fcbf8f9aaa573d",
		Date:       "2023-12-03T01:20:59Z",
		Signature:  "5ofg7njFDTrn9K2KfMVkg8JF8C12TU23dXMb2EZyiqPW",
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
