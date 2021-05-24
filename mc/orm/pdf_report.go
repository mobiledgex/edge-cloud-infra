package orm

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/wcharczuk/go-chart/v2"
)

const (
	FontName            = "Arial"
	DefaultFontSize     = float64(10)
	ReportTitleFontSize = float64(20)
	PageTitleFontSize   = float64(15)
	HeaderFontSize      = float64(8)
	TitleLogoSize       = float64(60)
	ChartWidth          = 150
	ChartHeight         = 60
	SectionGap          = float64(15)
)

type AxisValueFormat int

const (
	AxisValueFormatInt AxisValueFormat = iota
	AxisValueFormatFloat
)

type ChartSpec struct {
	FillColor        bool
	ShowLegend       bool
	YAxisValueFormat AxisValueFormat
}

type TimeChartData struct {
	Name    string
	XValues []time.Time
	YValues []float64
}
type TimeChartDataMap map[string][]TimeChartData

type PieChartDataMap map[string]float64

type CellInfo struct {
	List   [][]byte
	Height float64
}

type PDFReport struct {
	timezone *time.Location
	pdf      *gofpdf.Fpdf
}

func NewReport(report *ormapi.GenerateReport) (*PDFReport, error) {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.SetFont(FontName, "B", DefaultFontSize)
	pdf.AliasNbPages("")
	location, err := time.LoadLocation(report.Timezone)
	if err != nil {
		return nil, err
	}
	pdfReport := &PDFReport{
		pdf:      pdf,
		timezone: location,
	}
	return pdfReport, nil
}

func (r *PDFReport) AddReportTitle(logoPath string) {
	r.pdf.SetFont(FontName, "B", ReportTitleFontSize)
	_, topMargin, rightMargin, _ := r.pdf.GetMargins()
	r.pdf.Cell(100, 10, "Cloudlet Usage Report")
	pageW, _ := r.pdf.GetPageSize()
	// Logo aspect ratio is 6:1
	r.pdf.ImageOptions(logoPath, pageW-rightMargin-60, topMargin, TitleLogoSize, TitleLogoSize/6, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
	r.pdf.Ln(-1)
}

func (r *PDFReport) AddPageTitle(titleName string) {
	r.pdf.SetFont(FontName, "B", PageTitleFontSize)
	r.pdf.Cell(100, 10, titleName)
	r.pdf.Ln(-1)
}

func (r *PDFReport) ResetHeader() {
	r.pdf.SetHeaderFunc(nil)
}

func (r *PDFReport) AddHeader(report *ormapi.GenerateReport, logoPath, cloudlet string) {
	r.pdf.SetHeaderFuncMode(func() {
		r.pdf.SetFont(FontName, "I", HeaderFontSize)
		headerStr := ""
		if cloudlet == "" {
			headerStr = fmt.Sprintf("Operator: %s | Region: %s", report.Org, report.Region)
		} else {
			headerStr = fmt.Sprintf("Operator: %s | Region: %s | Cloudlet: %s", report.Org, report.Region, cloudlet)
		}
		_, topMargin, rightMargin, _ := r.pdf.GetMargins()
		r.pdf.CellFormat(0, 0, headerStr, "", 0, "L", false, 0, "")
		pageW, _ := r.pdf.GetPageSize()
		// Logo aspect ratio is 6:1
		r.pdf.ImageOptions(logoPath, pageW-rightMargin-24, topMargin-2, 24, 4, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
		r.pdf.Ln(5)
		r.AddHorizontalLine()
		r.pdf.Ln(5)
	}, false)
}

func (r *PDFReport) AddFooter() {
	r.pdf.SetFooterFunc(func() {
		r.pdf.SetY(-15)
		r.pdf.SetFont(FontName, "I", HeaderFontSize)
		r.pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", r.pdf.PageNo()), "", 0, "C", false, 0, "")
	})
}

func (r *PDFReport) AddOperatorInfo(report *ormapi.GenerateReport) {
	r.pdf.SetFont(FontName, "B", HeaderFontSize)
	r.pdf.Cell(40, 10, fmt.Sprintf("Operator: %s", report.Org))
	r.pdf.Ln(5)
	r.pdf.Cell(40, 10, fmt.Sprintf("Region: %s", report.Region))
	r.pdf.Ln(5)
	startDate := report.StartTimeUTC.In(r.timezone).Format(ormapi.TimeFormatDate)
	endDate := report.EndTimeUTC.In(r.timezone).Format(ormapi.TimeFormatDate)
	r.pdf.Cell(40, 10, fmt.Sprintf("Report Period: %s - %s", startDate, endDate))
	r.pdf.Ln(5)
	r.pdf.Cell(40, 10, fmt.Sprintf("Timezone: %s", report.Timezone))
	r.pdf.Ln(-1)
}

func (r *PDFReport) AddHorizontalLine() {
	pageW, _ := r.pdf.GetPageSize()
	_, leftMargin, rightMargin, _ := r.pdf.GetMargins()
	curY := r.pdf.GetY()
	r.pdf.Line(leftMargin, curY, pageW-rightMargin, curY)
	r.pdf.DrawPath("DF")
}

func (r *PDFReport) AddPage() {
	r.pdf.AddPage()
}

func (r *PDFReport) Err() error {
	if r.pdf.Err() {
		return r.pdf.Error()
	}
	return nil
}

func (r *PDFReport) Output(buf *bytes.Buffer) error {
	if buf == nil {
		return fmt.Errorf("Empty buffer")
	}
	return r.pdf.Output(buf)
}

func (r *PDFReport) AddTable(title string, hdr []string, tbl [][]string, colsWidth []float64) {
	rowHeight := float64(7)
	cellGap := float64(2)

	if len(tbl) == 0 {
		// no data, skip adding table
		return
	}

	// Center align all columns
	alignCols := []string{}
	for _, _ = range hdr {
		alignCols = append(alignCols, "C")
	}

	leftMargin, topMargin, _, bottomMargin := r.pdf.GetMargins()
	_, pageH := r.pdf.GetPageSize()
	pageBreakTrigger := pageH - bottomMargin
	maxRowHeight := rowHeight

	curY := r.pdf.GetY()
	if curY+maxRowHeight+rowHeight+cellGap+cellGap+10 > pageBreakTrigger {
		// break, go to new page
		r.pdf.AddPage()
		curY = topMargin + 10
	}

	// Title
	r.pdf.SetFont(FontName, "B", DefaultFontSize)
	r.pdf.Cell(40, 10, title)
	r.pdf.Ln(-1)

	// Table Header
	r.pdf.SetFont(FontName, "B", DefaultFontSize)
	r.pdf.SetFillColor(240, 240, 240)
	for ii, str := range hdr {
		r.pdf.CellFormat(colsWidth[ii], rowHeight, str, "1", 0, alignCols[ii], true, 0, "")
	}
	r.pdf.Ln(-1)

	// Table Content
	r.pdf.SetFont(FontName, "", DefaultFontSize)
	r.pdf.SetFillColor(255, 255, 255)

	curY = r.pdf.GetY()
	for _, line := range tbl {
		// Cell height calculation loop
		cellList := []CellInfo{}
		for colIndex, colStr := range line {
			cell := CellInfo{}
			cell.List = r.pdf.SplitLines([]byte(colStr), colsWidth[colIndex]-cellGap-cellGap)
			cell.Height = float64(len(cell.List)) * rowHeight
			cellList = append(cellList, cell)
			if cell.Height > maxRowHeight {
				maxRowHeight = cell.Height
			}
		}
		// Cell render loop
		curX := leftMargin
		if curY+maxRowHeight+cellGap+cellGap > pageBreakTrigger {
			// break, go to new page
			r.pdf.AddPage()
			curY = topMargin + 10
		}
		for colIndex, _ := range line {
			r.pdf.Rect(curX, curY, colsWidth[colIndex], maxRowHeight+cellGap+cellGap, "D")
			cell := cellList[colIndex]
			cellY := curY + cellGap + (maxRowHeight-cell.Height)/2
			for splitNo := 0; splitNo < len(cell.List); splitNo++ {
				r.pdf.SetXY(curX+cellGap, cellY)
				r.pdf.CellFormat(colsWidth[colIndex]-cellGap-cellGap, rowHeight, string(cell.List[splitNo]), "", 0, alignCols[colIndex], false, 0, "")
				cellY += rowHeight
			}
			curX += colsWidth[colIndex]
		}
		curY += maxRowHeight + cellGap + cellGap
	}
	r.pdf.Ln(SectionGap)
}

// Adds TimeCharts for resource data in the format: map<resource>[]resourcedata
func (r *PDFReport) AddResourceTimeCharts(chartPrefix string, charts TimeChartDataMap, spec ChartSpec) error {
	keys := []string{}
	for k := range charts {
		keys = append(keys, k)
	}
	// sort chart data
	sort.Strings(keys)

	for _, key := range keys {
		multiChartData := charts[key]
		render := false
		for _, chartData := range multiChartData {
			if len(chartData.XValues) < 2 {
				// there should be atleast 2 points, skip chart
				continue
			}
			allZeroes := true
			for _, yVal := range chartData.YValues {
				if yVal != 0 {
					allZeroes = false
					break
				}
			}
			if allZeroes {
				// no data to render, skip chart
				continue
			}
			render = true
		}
		if !render {
			continue
		}

		err := r.AddTimeChart(chartPrefix, key, charts[key], spec)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PDFReport) setChartPageBreak(chartHeight float64) {
	_, topMargin, _, bottomMargin := r.pdf.GetMargins()
	_, pageH := r.pdf.GetPageSize()
	pageBreakTrigger := pageH - bottomMargin

	curY := r.pdf.GetY()
	if curY+chartHeight+10 > pageBreakTrigger {
		// break, go to new page
		r.pdf.AddPage()
		curY = topMargin + 10
	}
}

// TimeValueFormatterWithFormatTZ returns a time formatter with a
// given format along with timezone
func TimeValueFormatterWithFormatTZ(dateFormat string, timezone *time.Location) chart.ValueFormatter {
	return func(v interface{}) string {
		if timezone == nil {
			return ""
		}
		if typed, isTyped := v.(time.Time); isTyped {
			return typed.In(timezone).Format(dateFormat)
		}
		if typed, isTyped := v.(int64); isTyped {
			return time.Unix(0, typed).In(timezone).Format(dateFormat)
		}
		if typed, isTyped := v.(float64); isTyped {
			return time.Unix(0, int64(typed)).In(timezone).Format(dateFormat)
		}
		return ""
	}
}

func (r *PDFReport) AddTimeChart(chartPrefix, title string, multiChartData []TimeChartData, spec ChartSpec) error {
	allChartSeries := []chart.Series{}
	maxVal := float64(10)
	for _, chartData := range multiChartData {
		for _, yVal := range chartData.YValues {
			if yVal > maxVal {
				maxVal = yVal
			}
		}

		chartSeries := CustomTimeSeries{
			chart.TimeSeries{
				Name:    chartData.Name,
				XValues: chartData.XValues,
				YValues: chartData.YValues,
			},
		}
		if spec.FillColor {
			chartSeries.Style.FillColor = chart.ColorBlue.WithAlpha(100)
			chartSeries.Style.StrokeColor = chart.ColorBlue
		}

		maxSeries := &chart.MaxSeries{
			Style: chart.Style{
				StrokeColor:     chart.ColorAlternateGray,
				StrokeDashArray: []float64{5.0, 5.0},
			},
			InnerSeries: chartSeries,
		}
		allChartSeries = append(allChartSeries, chartSeries)
		allChartSeries = append(allChartSeries, maxSeries)
	}

	r.setChartPageBreak(float64(ChartHeight))

	r.pdf.SetFont(FontName, "B", DefaultFontSize)
	titleWords := util.SplitCamelCase(title)
	title = strings.Title(strings.Join(titleWords, " "))
	r.pdf.Cell(40, 10, title)
	r.pdf.Ln(-1)

	graph := chart.Chart{
		YAxis: chart.YAxis{
			Range: &chart.ContinuousRange{
				Min: 0,
			},
		},
		XAxis: chart.XAxis{
			ValueFormatter: TimeValueFormatterWithFormatTZ(ormapi.TimeFormatDateTime, r.timezone),
		},
		Series: allChartSeries,
	}

	if spec.YAxisValueFormat == AxisValueFormatInt {
		graph.YAxis.ValueFormatter = chart.IntValueFormatter
		graph.YAxis.Range = &chart.ContinuousRange{
			Max: maxVal,
		}
	}

	if spec.ShowLegend {
		//note we have to do this as a separate step because we need a reference to graph
		graph.Elements = []chart.Renderable{
			chart.Legend(&graph),
		}
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		return fmt.Errorf("failed rendering graph: %s", err.Error())
	}
	// unique for the chart image
	imgName := chartPrefix + title
	return r.DrawChart(imgName, ChartWidth, ChartHeight, buffer)
}

func (r *PDFReport) AddPieChart(cloudletName, title string, data PieChartDataMap) error {
	chartValues := []chart.Value{}

	keys := []string{}
	for key, _ := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	allZeroes := true
	maxVal := float64(10)
	for _, label := range keys {
		value := data[label]
		chartValues = append(chartValues, chart.Value{
			Value: value,
			Label: fmt.Sprintf("%s (%.0f)", label, value),
		})
		if value != 0 {
			allZeroes = false
		}
		if value > maxVal {
			maxVal = value
		}
	}

	if allZeroes {
		return nil
	}

	chartWidth := 80
	chartHeight := chartWidth
	r.setChartPageBreak(float64(chartHeight))

	r.pdf.SetFont(FontName, "B", DefaultFontSize)
	r.pdf.Cell(40, 10, title)
	r.pdf.Ln(-1)

	graph := chart.PieChart{
		Values: chartValues,
		Background: chart.Style{
			Padding: chart.Box{
				Left:  30,
				Right: 30,
			},
		},
		SliceStyle: chart.Style{
			FontSize: 20,
		},
	}
	if len(chartValues) == 1 {
		graph.SliceStyle.FontColor = chart.ColorAlternateBlue
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		return fmt.Errorf("failed rendering graph: %s", err.Error())
	}
	imgName := cloudletName + strings.Replace(title, " ", "_", -1)
	return r.DrawChart(imgName, chartWidth, chartHeight, buffer)
}

func (r *PDFReport) DrawChart(imgName string, chartWidth, chartHeight int, buffer *bytes.Buffer) error {
	imgOpts := gofpdf.ImageOptions{
		ReadDpi:   false,
		ImageType: "PNG",
	}
	r.pdf.RegisterImageOptionsReader(imgName, imgOpts, buffer)
	curX, curY := r.pdf.GetXY()
	r.pdf.ImageOptions(imgName, curX, curY, float64(chartWidth), float64(chartHeight), true, imgOpts, 0, "")
	r.pdf.Ln(SectionGap)
	return nil
}

// CustomTimeSeries implements new Render function to add
// a step based interpolation for accurate usage representation
type CustomTimeSeries struct {
	chart.TimeSeries
}

// Render renders the series.
func (ts CustomTimeSeries) Render(r chart.Renderer, canvasBox chart.Box, xrange, yrange chart.Range, defaults chart.Style) {
	style := ts.Style.InheritFrom(defaults)
	LineSeries(r, canvasBox, xrange, yrange, style, ts)
}

// LineSeries draws a line series with a step based interpolation.
func LineSeries(r chart.Renderer, canvasBox chart.Box, xrange, yrange chart.Range, style chart.Style, vs chart.ValuesProvider) {
	if vs.Len() == 0 {
		return
	}

	cb := canvasBox.Bottom
	cl := canvasBox.Left

	v0x, v0y := vs.GetValues(0)
	x0 := cl + xrange.Translate(v0x)
	y0 := cb - yrange.Translate(v0y)

	yv0 := yrange.Translate(0)

	var curVy, newVx, newVy float64
	var curY, newX, newY int

	if style.ShouldDrawStroke() && style.ShouldDrawFill() {
		style.GetFillOptions().WriteDrawingOptionsToRenderer(r)
		r.MoveTo(x0, y0)
		for i := 1; i < vs.Len(); i++ {
			_, curVy = vs.GetValues(i - 1)
			newVx, newVy = vs.GetValues(i)
			curY = cb - yrange.Translate(curVy)
			newX = cl + xrange.Translate(newVx)
			newY = cb - yrange.Translate(newVy)
			r.LineTo(newX, curY)
			r.LineTo(newX, newY)
		}
		r.LineTo(newX, chart.MinInt(cb, cb-yv0))
		r.LineTo(x0, chart.MinInt(cb, cb-yv0))
		r.LineTo(x0, y0)
		r.Fill()
	}

	if style.ShouldDrawStroke() {
		style.GetStrokeOptions().WriteDrawingOptionsToRenderer(r)

		r.MoveTo(x0, y0)
		for i := 1; i < vs.Len(); i++ {
			_, curVy = vs.GetValues(i - 1)
			newVx, newVy = vs.GetValues(i)
			curY = cb - yrange.Translate(curVy)
			newX = cl + xrange.Translate(newVx)
			newY = cb - yrange.Translate(newVy)
			r.LineTo(newX, curY)
			r.LineTo(newX, newY)
		}
		r.Stroke()
	}
}
