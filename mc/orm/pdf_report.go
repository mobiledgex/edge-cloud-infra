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
	FontName              = "Arial"
	TimeFormatDate        = "2006/01/02"
	TimeFormatDateTime    = "01-02 15:04:05"
	TimeFormatDayDateTime = "Mon Jan 2 15:04:05"

	DefaultFontSize     = float64(10)
	ReportTitleFontSize = float64(20)
	PageTitleFontSize   = float64(15)
	HeaderFontSize      = float64(8)
	TitleLogoSize       = float64(60)
	ChartWidth          = 150
	ChartHeight         = 60
	SectionGap          = float64(15)
)

type TimeChartData struct {
	Name    string
	XValues []time.Time
	YValues []float64
}

type BarChartData struct {
	Name    string
	XValues []string
	YValues []float64
}

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

func (r *PDFReport) AddHeader(report *ormapi.GenerateReport, logoPath, cloudlet string) {
	r.pdf.SetHeaderFuncMode(func() {
		r.pdf.SetFont(FontName, "I", HeaderFontSize)
		headerStr := ""
		if cloudlet == "" {
			headerStr = fmt.Sprintf("Operator: %s | Region: %s | Timezone: %s", report.Org, report.Region, report.Timezone)
		} else {
			headerStr = fmt.Sprintf("Operator: %s | Region: %s | Timezone: %s | Cloudlet: %s", report.Org, report.Region, report.Timezone, cloudlet)
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
	startDate := report.StartTime.Format(TimeFormatDate)
	endDate := report.EndTime.Format(TimeFormatDate)
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

func (r *PDFReport) Save(filepath string) error {
	return r.pdf.OutputFileAndClose(filepath)
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

func (r *PDFReport) AddTimeCharts(charts map[string]TimeChartData) error {
	keys := []string{}
	for k := range charts {
		chartData := charts[k]
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
		keys = append(keys, k)
	}

	if len(keys) == 0 {
		return nil
	}

	// sort chart data
	sort.Strings(keys)

	for _, key := range keys {
		err := r.AddTimeChart(key, charts[key])
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PDFReport) AddPieChart(title string, data BarChartData) error {
	chartValues := []chart.Value{}
	allZeroes := true
	maxVal := float64(10)
	for ii, _ := range data.XValues {
		yVal := data.YValues[ii]
		xVal := data.XValues[ii]
		chartValues = append(chartValues, chart.Value{
			Value: yVal,
			Label: fmt.Sprintf("%s (%.0f)", xVal, yVal),
		})
		if yVal != 0 {
			allZeroes = false
		}
		if yVal > maxVal {
			maxVal = yVal
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

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		return fmt.Errorf("failed rendering graph: %s", err.Error())
	}
	return r.DrawChart(data.Name, chartWidth, chartHeight, buffer)
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

func (r *PDFReport) AddTimeChart(title string, chartData TimeChartData) error {
	maxVal := float64(10)
	for ii, _ := range chartData.XValues {
		xVal := chartData.XValues[ii]
		yVal := chartData.YValues[ii]
		if yVal > maxVal {
			maxVal = yVal
		}
	}

	chartSeries := CustomTimeSeries{
		chart.TimeSeries{
			Style: chart.Style{
				StrokeColor: chart.ColorBlue,
				FillColor:   chart.ColorBlue.WithAlpha(100),
			},
			XValues: chartData.XValues,
			YValues: chartData.YValues,
		},
	}

	maxSeries := &chart.MaxSeries{
		Style: chart.Style{
			StrokeColor:     chart.ColorAlternateGray,
			StrokeDashArray: []float64{5.0, 5.0},
		},
		InnerSeries: chartSeries,
	}

	r.setChartPageBreak(float64(ChartHeight))

	r.pdf.SetFont(FontName, "B", DefaultFontSize)
	titleWords := util.SplitCamelCase(title)
	title = strings.Title(strings.Join(titleWords, " "))
	r.pdf.Cell(40, 10, title)
	r.pdf.Ln(-1)

	graph := chart.Chart{
		YAxis: chart.YAxis{
			ValueFormatter: chart.IntValueFormatter,
			Range: &chart.ContinuousRange{
				Max: maxVal,
			},
		},
		XAxis: chart.XAxis{
			ValueFormatter: TimeValueFormatterWithFormatTZ(TimeFormatDateTime, r.timezone),
		},
		Series: []chart.Series{
			chartSeries,
			maxSeries,
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		return fmt.Errorf("failed rendering graph: %s", err.Error())
	}
	imgName := title + chartData.Name
	return r.DrawChart(imgName, ChartWidth, ChartHeight, buffer)
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
