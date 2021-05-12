package ormapi

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

var reportFileNameTests = map[string]string{
	"TDG_20210420_20210503_TDGReporter_report.pdf":                           "TDG",
	"TDG_11111111_11111111_20210420_20210503_test_report.pdf":                "TDG_11111111_11111111",
	"TDG_11111111_11111111_xyz_report.pdf_20210420_20210503_test_report.pdf": "TDG_11111111_11111111_xyz_report.pdf",
}

func TestReportFileNameRegex(t *testing.T) {
	regObj := regexp.MustCompile(GetReportFileNameRE())
	for inp, out := range reportFileNameTests {
		allStrs := regObj.FindStringSubmatch(inp)
		require.Greater(t, len(allStrs), 1)
		orgName := allStrs[1]
		require.Equal(t, orgName, out)
	}
}
