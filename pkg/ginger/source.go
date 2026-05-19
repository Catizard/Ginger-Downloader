package ginger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/rotisserie/eris"
)

type DownloadSourceMeta struct {
	Name               string
	MetaMD5QueryURL    string
	MetaSha256QueryURL string
}

type DownloadSource interface {
	GetDownloadURLFromMD5(string) (DownloadInfo, error)
	GetDownloadURLFromSha256(string) (DownloadInfo, error)
}

type DownloadInfo struct {
	DownloadURL  string
	UniqueSymbol string
	FileName     string
}

var _ DownloadSource = (*gingerDownloadSource)(nil)

var GingerDownloadSource gingerDownloadSource = gingerDownloadSource{
	Meta: DownloadSourceMeta{
		Name:               "ginger",
		MetaMD5QueryURL:    fmt.Sprintf("%s%s", SERVER_BASE_URL, QUERY_PACKAGE_MD5_API),
		MetaSha256QueryURL: fmt.Sprintf("%s%s", SERVER_BASE_URL, QUERY_PACKAGE_SHA256_API),
	},
}

type gingerDownloadSource struct {
	Meta DownloadSourceMeta
}

func (d *gingerDownloadSource) GetMeta() DownloadSourceMeta {
	return d.Meta
}

func (d *gingerDownloadSource) GetDownloadURLFromMD5(md5 string) (downloadInfo DownloadInfo, err error) {
	metaQueryURL := fmt.Sprintf("%s%s", d.Meta.MetaMD5QueryURL, md5)
	resp, err := d.queryPackage(metaQueryURL)
	if err != nil {
		return DownloadInfo{}, err
	}
	return DownloadInfo{
		DownloadURL:  resp.DownloadURL,
		UniqueSymbol: resp.DownloadURL,
		FileName:     resp.FileName,
	}, nil
}

func (d *gingerDownloadSource) GetDownloadURLFromSha256(sha256 string) (downloadInfo DownloadInfo, err error) {
	metaQueryURL := fmt.Sprintf("%s%s", d.Meta.MetaSha256QueryURL, sha256)
	resp, err := d.queryPackage(metaQueryURL)
	if err != nil {
		return DownloadInfo{}, err
	}
	return DownloadInfo{
		DownloadURL:  resp.DownloadURL,
		UniqueSymbol: resp.DownloadURL,
		FileName:     resp.FileName,
	}, nil
}

func (d *gingerDownloadSource) queryPackage(metaQueryURL string) (*mResp, error) {
	log.Printf("Querying package: %s", metaQueryURL)
	resp, err := http.Get(metaQueryURL)
	if err != nil {
		return nil, eris.Wrap(err, "get meta")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, eris.Errorf("error code: %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, eris.Wrap(err, "http read")
	}
	if strings.HasPrefix(string(b), "404") {
		return nil, eris.Errorf("404 NOT FOUND")
	}
	var result mResp
	if err = json.Unmarshal(b, &result); err != nil {
		return nil, eris.Wrapf(err, "failed to unmarshal result: %s", string(b))
	}
	return &result, nil
}

func (d *gingerDownloadSource) AllowBatchDownload() bool {
	return true
}

// Ginger server models
type mResp struct {
	ShardMD5    string `json:"shardMD5"`
	FileName    string `json:"fileName"`
	FileSize    int64  `json:"fileSize"`
	DirectoryID string `json:"directoryID"`
	MD5s        string `json:"md5s"`
	DownloadURL string `json:"downloadURL"`
}
