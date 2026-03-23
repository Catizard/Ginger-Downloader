package ginger_test

import (
	"fmt"
	"testing"

	"github.com/Catizard/Ginger-Downloader/pkg/ginger"
)

func TestQueryTableSummary(t *testing.T) {
	headers, err := ginger.QueryTableSummary()
	if err != nil {
		t.Errorf("QueryTableSummary: %s", err)
	}
	for i, header := range headers {
		fmt.Printf("%d -> %s(%d/%d)\n", i, header.HeaderURL, header.DataCount-header.MissingCount, header.DataCount)
	}
}
