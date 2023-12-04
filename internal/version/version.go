// Code generated by the go generate ./... command; DO NOT EDIT.

package version

import "fmt"

var ver *Version

func init() {
	ver = &Version{
		Version:    "v0.0.1-dev",
		CommitHash: "27882c594d3a805ddf2bef1d3834b215c5d895f9",
		Date:       "2023-12-03T23:41:56Z",
		Signature:  "F8Vs5swHjDReTw13v2DANXxy6u3TDXNAxebGeEwHmiXu",
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