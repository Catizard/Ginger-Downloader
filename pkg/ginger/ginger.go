// Package ginger: Providing integration with ginger rush server
package ginger

import (
	req "github.com/imroc/req/v3"
	"github.com/rotisserie/eris"
)

const (
	SERVER_BASE_URL          = "https://gingerrush.com/api/v1/"
	QUERY_PACKAGE_MD5_API    = "files/package/"
	QUERY_PACKAGE_SHA256_API = "files/package_sha256/"
)

var client = req.NewClient().SetBaseURL(SERVER_BASE_URL)

type TableHeader struct {
	HeaderURL    string
	Name         string
	Symbol       string
	DataCount    int
	MissingCount int
}

type ServerInfoSummary struct {
	Headers []TableHeader
}

func Initialize() (*ServerInfoSummary, error) {
	headers, err := QueryTableSummary()
	if err != nil {
		return nil, err
	}
	return &ServerInfoSummary{
		Headers: headers,
	}, nil
}

func QueryTableSummary() ([]TableHeader, error) {
	var ret []TableHeader
	resp, err := client.R().
		SetBody(struct{}{}).
		SetSuccessResult(&ret).Post("table/selectHeaderList")
	if err != nil {
		return nil, eris.Wrapf(err, "bad http request")
	}
	if !resp.IsSuccessState() {
		return nil, eris.Errorf("bad request status: %d", resp.StatusCode)
	}
	return ret, nil
}
