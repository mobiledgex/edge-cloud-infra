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

const (
	FontName     = "Arial"
	LogoFilePath = "/Users/ashishjain/Desktop/MobiledgeX_Logo.png"

	TimeFormatDate     = "2006/01/02"
	TimeFormatDateTime = "2006-01-02T15:04:05Z"

	DefaultFontSize    = float64(10)
	TitleFontSize      = float64(20)
	HeaderFontSize     = float64(8)
	TitleLogoSize      = float64(60)
	ChartWidth         = float64(150)
	ChartHeight        = float64(60)
	SectionGap         = float64(15)
	DefaultColumnWidth = float64(50)
)

type TimeChartData struct {
	Name    string
	XValues []time.Time
	YValues []float64
}

type CellInfo struct {
	List   [][]byte
	Height float64
}

func NewReport() *gofpdf.Fpdf {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.SetFont(FontName, "B", DefaultFontSize)
	pdf.AliasNbPages("")
	return pdf
}

func AddPageTitle(pdf *gofpdf.Fpdf) {
	pdf.SetFont(FontName, "B", TitleFontSize)
	_, topMargin, rightMargin, _ := pdf.GetMargins()
	pdf.Cell(100, 10, "Cloudlet Usage Report")
	pageW, _ := pdf.GetPageSize()
	// Logo aspect ratio is 6:1
	pdf.ImageOptions(LogoFilePath, pageW-rightMargin-60, topMargin, TitleLogoSize, TitleLogoSize/6, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
	pdf.Ln(-1)
}

func AddHeader(pdf *gofpdf.Fpdf, report *ormapi.GenerateReport) {
	pdf.SetHeaderFuncMode(func() {
		pdf.SetFont(FontName, "I", HeaderFontSize)
		headerStr := fmt.Sprintf("Operator: %s | Region: %s", report.Org, report.Region)
		_, topMargin, rightMargin, _ := pdf.GetMargins()
		pdf.CellFormat(0, 0, headerStr, "", 0, "L", false, 0, "")
		pageW, _ := pdf.GetPageSize()
		// Logo aspect ratio is 6:1
		pdf.ImageOptions(LogoFilePath, pageW-rightMargin-24, topMargin-2, 24, 4, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
		pdf.Ln(5)
		AddHorizontalLine(pdf)
		pdf.Ln(5)
	}, false)
}

func AddFooter(pdf *gofpdf.Fpdf) {
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont(FontName, "I", HeaderFontSize)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})
}

func AddOperatorInfo(pdf *gofpdf.Fpdf, report *ormapi.GenerateReport) {
	pdf.SetFont(FontName, "B", HeaderFontSize)
	pdf.Cell(40, 10, fmt.Sprintf("Operator: %s", report.Org))
	pdf.Ln(5)
	pdf.Cell(40, 10, fmt.Sprintf("Region: %s", report.Region))
	pdf.Ln(5)
	startDate := report.StartTime.Format(TimeFormatDate)
	endDate := report.EndTime.Format(TimeFormatDate)
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

func AddTable(pdf *gofpdf.Fpdf, title string, hdr []string, tbl [][]string, colWidth float64) {
	if colWidth <= 0 {
		colWidth = DefaultColumnWidth
	}
	rowHeight := float64(7)
	cellGap := float64(2)

	pdf.SetFont(FontName, "B", DefaultFontSize)
	pdf.Cell(40, 10, title)
	pdf.Ln(-1)

	// Center align all columns
	alignCols := []string{}
	for _, _ = range hdr {
		alignCols = append(alignCols, "C")
	}

	// Table Header
	pdf.SetFont(FontName, "B", DefaultFontSize)
	pdf.SetFillColor(240, 240, 240)
	for ii, str := range hdr {
		pdf.CellFormat(colWidth, rowHeight, str, "1", 0, alignCols[ii], true, 0, "")
	}
	pdf.Ln(-1)

	// Table Content
	pdf.SetFont(FontName, "", DefaultFontSize)
	pdf.SetFillColor(255, 255, 255)

	maxRowHeight := rowHeight
	_, leftMargin, _, _ := pdf.GetMargins()
	curY := pdf.GetY()
	for _, line := range tbl {
		// Cell height calculation loop
		cellList := []CellInfo{}
		for _, colStr := range line {
			cell := CellInfo{}
			cell.List = pdf.SplitLines([]byte(colStr), colWidth-cellGap-cellGap)
			cell.Height = float64(len(cell.List)) * rowHeight
			cellList = append(cellList, cell)
			if cell.Height > maxRowHeight {
				maxRowHeight = cell.Height
			}
		}
		// Cell render loop
		curX := leftMargin
		for colNo, _ := range line {
			pdf.Rect(curX, curY, colWidth, maxRowHeight+cellGap+cellGap, "D")
			cell := cellList[colNo]
			cellY := curY + cellGap + (maxRowHeight-cell.Height)/2
			for splitNo := 0; splitNo < len(cell.List); splitNo++ {
				pdf.SetXY(curX+cellGap, cellY)
				pdf.CellFormat(colWidth-cellGap-cellGap, rowHeight, string(cell.List[splitNo]), "", 0, alignCols[colNo], false, 0, "")
				cellY += rowHeight
			}
			curX += colWidth
		}
		curY += maxRowHeight + cellGap + cellGap
	}
	pdf.Ln(SectionGap)
}

func AddTimeCharts(pdf *gofpdf.Fpdf, title string, charts map[string][]TimeChartData) error {
	pdf.SetFont(FontName, "B", DefaultFontSize)
	pdf.Cell(40, 10, title)
	pdf.Ln(-1)
	// sort chart data
	keys := []string{}
	for k := range charts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		chart := charts[key]
		sort.Slice(chart[:], func(i, j int) bool {
			return chart[i].Name < chart[j].Name
		})
		err := AddTimeChart(pdf, key, chart)
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
			ValueFormatter: chart.TimeValueFormatterWithFormat(TimeFormatDateTime),
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
	pdf.ImageOptions(imgName, curX, curY, ChartWidth, ChartHeight, true, imgOpts, 0, "")
	pdf.Ln(SectionGap)
	return nil
}
