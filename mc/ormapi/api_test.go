package ormapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var reportFileNameTests = map[string][]string{
	"TDG/TDGReporter/20210420_20210503.pdf":                           []string{"TDG", "TDGReporter"},
	"TDG_11111111_11111111/test/20210420_20210503.pdf":                []string{"TDG_11111111_11111111", "test"},
	"TDG_11111111_11111111_xyz_report.pdf/test/20210420_20210503.pdf": []string{"TDG_11111111_11111111_xyz_report.pdf", "test"},
}

func TestReportFileName(t *testing.T) {
	for inp, out := range reportFileNameTests {
		orgName, reporterName := GetInfoFromReportFileName(inp)
		require.Equal(t, orgName, out[0])
		require.Equal(t, reporterName, out[1])
	}
}
