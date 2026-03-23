package ginger

import "github.com/rotisserie/eris"

type SabunHash struct {
	SHA256 string
	MD5    string
}

type DownloadContext struct {
	header         *TableHeader
	ignoringHashes []SabunHash
}

func NewDownloadContext() *DownloadContext {
	return &DownloadContext{}
}

func (c *DownloadContext) Header(header *TableHeader) *DownloadContext {
	c.header = header
	return c
}

func (c *DownloadContext) IgnoringHashes(ignoringHashes []SabunHash) *DownloadContext {
	c.ignoringHashes = ignoringHashes
	return c
}

func (c *DownloadContext) Start() error {
	if c.header == nil {
		return eris.Errorf("no header provided")
	}
	// TODO: Implement me!
	return nil
}

// TODO: Expose the download process here
