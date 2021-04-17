package orm

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/wcharczuk/go-chart/v2"
)

func NewReport() *gofpdf.Fpdf {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.AliasNbPages("")
	return pdf
}

func AddPageTitle(pdf *gofpdf.Fpdf) {
	pdf.SetFont("Arial", "B", 20)
	_, topMargin, rightMargin, _ := pdf.GetMargins()
	pdf.Cell(100, 10, "Cloudlet Usage Report")
	pageW, _ := pdf.GetPageSize()
	// Logo aspect ratio is 6:1
	pdf.ImageOptions("/Users/ashishjain/Desktop/MobiledgeX_Logo.png", pageW-rightMargin-60, topMargin, 60, 10, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
	pdf.Ln(-1)
}

func AddHeader(pdf *gofpdf.Fpdf, report *ormapi.GenerateReport) {
	pdf.SetHeaderFuncMode(func() {
		pdf.SetFont("Arial", "I", 8)
		headerStr := fmt.Sprintf("Operator: %s | Region: %s", report.Org, report.Region)
		_, topMargin, rightMargin, _ := pdf.GetMargins()
		pdf.CellFormat(0, 0, headerStr, "", 0, "L", false, 0, "")
		pageW, _ := pdf.GetPageSize()
		// Logo aspect ratio is 6:1
		pdf.ImageOptions("/Users/ashishjain/Desktop/MobiledgeX_Logo.png", pageW-rightMargin-24, topMargin-2, 24, 4, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")

		pdf.Ln(5)
		AddHorizontalLine(pdf)
		pdf.Ln(5)
	}, false)
}

func AddFooter(pdf *gofpdf.Fpdf) {
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})
}

func AddOperatorInfo(pdf *gofpdf.Fpdf, report *ormapi.GenerateReport) {
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(40, 10, fmt.Sprintf("Operator: %s", report.Org))
	pdf.Ln(5)
	pdf.Cell(40, 10, fmt.Sprintf("Region: %s", report.Region))
	pdf.Ln(5)
	startDate := report.StartTime.Format("2006/01/02")
	endDate := report.EndTime.Format("2006/01/02")
	pdf.Cell(40, 10, fmt.Sprintf("Report Period: %s - %s", startDate, endDate))
	pdf.Ln(-1)
}

func AddHorizontalLine(pdf *gofpdf.Fpdf) {
	pageW, _ := pdf.GetPageSize()
	_, leftMargin, rightMargin, _ := pdf.GetMargins()
	curY := pdf.GetY()
	pdf.Line(leftMargin, curY, pageW-rightMargin, curY)
	pdf.DrawPath("DF")
}

func AddTable(pdf *gofpdf.Fpdf, title string, hdr []string, tbl [][]string, alignCols []string, width float64) {
	if width <= 0 {
		width = 50
	}
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 10, title)
	pdf.Ln(-1)

	// Header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	for ii, str := range hdr {
		pdf.CellFormat(width, 7, str, "1", 0, alignCols[ii], true, 0, "")
	}
	pdf.Ln(-1)

	// Table Content
	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(255, 255, 255)
	for lineNo, line := range tbl {
		borderStr := "LRT"
		for _, str := range line {
			if str == "" {
				borderStr = "LR"
				break
			}
		}
		if lineNo == len(tbl)-1 {
			borderStr = "LRB"
		}
		for ii, str := range line {
			pdf.CellFormat(width, 7, str, borderStr, 0, alignCols[ii], false, 0, "")
		}
		pdf.Ln(-1)
	}
	pdf.Ln(-1)
}

func getEntriesFromBlocks(key string, dataBlocks ...[]string) [][]string {
	maxLen := 0
	for _, block := range dataBlocks {
		if len(block) > maxLen {
			maxLen = len(block)
		}
	}
	entries := [][]string{}
	for ii := 0; ii < maxLen; ii++ {
		entry := []string{}
		if ii == 0 {
			entry = append(entry, key)
		} else {
			entry = append(entry, "")
		}
		for _, block := range dataBlocks {
			if ii < len(block) {
				entry = append(entry, block[ii])
			} else {
				entry = append(entry, "")
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

type TimeChartData struct {
	Name    string
	XValues []time.Time
	YValues []float64
}

func AddTimeCharts(pdf *gofpdf.Fpdf, title string, charts map[string][]TimeChartData) error {
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 10, title)
	pdf.Ln(-1)
	// sort chart data
	keys := []string{}
	for k := range charts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		err := AddTimeChart(pdf, key, charts[key])
		if err != nil {
			return err
		}
	}
	return nil
}

func AddTimeChart(pdf *gofpdf.Fpdf, title string, data []TimeChartData) error {
	chartSeries := []chart.Series{}
	xGridLines := []chart.GridLine{}
	yGridLines := []chart.GridLine{}
	allZeroes := true
	for _, chartData := range data {
		cloudletSeries := chart.TimeSeries{
			Name:    chartData.Name,
			XValues: chartData.XValues,
			YValues: chartData.YValues,
		}
		chartSeries = append(chartSeries, cloudletSeries)
		for _, yval := range chartData.YValues {
			if yval != float64(0) {
				allZeroes = false
				break
			}
			yGridLines = append(yGridLines, chart.GridLine{Value: yval})
		}
		for _, xval := range chartData.XValues {
			xGridLines = append(xGridLines, chart.GridLine{Value: chart.TimeToFloat64(xval)})
		}
	}
	if allZeroes {
		// Skip chart
		return nil
	}

	graph := chart.Chart{
		YAxis: chart.YAxis{
			ValueFormatter: func(v interface{}) string {
				if vf, isFloat := v.(float64); isFloat {
					return fmt.Sprintf("%0.0f", vf)
				}
				return ""
			},
			GridMajorStyle: chart.Style{
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 0.2,
			},
			GridLines: yGridLines,
		},
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeValueFormatterWithFormat("2006-01-02T15:04:05Z"),
			GridMajorStyle: chart.Style{
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 0.2,
			},
			GridLines: xGridLines,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:  40,
				Left: 100,
			},
		},
		Series: chartSeries,
		Title:  title,
	}

	// note: we have to do this as a separate step because we need a reference to graph
	graph.Elements = []chart.Renderable{chart.LegendLeft(&graph)}
	return DrawChart(pdf, title, &graph)
}

func DrawChart(pdf *gofpdf.Fpdf, imgName string, graph *chart.Chart) error {
	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		return fmt.Errorf("failed rendering graph: %s", err.Error())
	}

	imgOpts := gofpdf.ImageOptions{
		ReadDpi:   false,
		ImageType: "PNG",
	}
	pdf.RegisterImageOptionsReader(imgName, imgOpts, buffer)
	curX, curY := pdf.GetXY()
	pdf.ImageOptions(imgName, curX, curY, float64(150), float64(60), true, imgOpts, 0, "")
	pdf.Ln(5)
	return nil
}
