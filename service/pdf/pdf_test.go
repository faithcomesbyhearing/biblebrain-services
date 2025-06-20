package pdf_test

import (
	"testing"

	pdf_service "biblebrain-services/service/pdf"

	"github.com/go-pdf/fpdf"
	"github.com/stretchr/testify/assert"
)

// TestConfiguration ensures that the default configuration values are set as expected.
func TestConfiguration(t *testing.T) {
	t.Parallel()
	options := pdf_service.Configuration()

	assert.Equal(t, "A4", options.PageDimensions)
	assert.Equal(t, "P", options.PageLayout)
	assert.Equal(t, "Arial", options.FontFamily)
	assert.InDelta(t, 8.0, options.FontSize, 0)
	assert.Equal(t, "mm", options.PageUnits)
	assert.Equal(t, 4, options.CardsPerPage)
	assert.InDelta(t, 8.0, options.PageMargin, 0)
	assert.InDelta(t, 210.0, options.PageWidth, 0)
	assert.InDelta(t, 297.0, options.PageHeight, 0)
	assert.InDelta(t, 99.25, options.CardWidth, 0)
	assert.InDelta(t, 1.0, options.CardPadding, 0)
	assert.Equal(t, "0", options.BorderText)
	assert.InDelta(t, 140.25, options.CardHeight, 0.25)
	assert.InDelta(t, 20.0, options.ImgHeightMax, 0)
	assert.InDelta(t, 70.0, options.ImgWidthMax, 0)
	assert.InDelta(t, 10.0, options.CellHeight, 0)
}

// TestCalculateOrgInfoHeight validates the calculation of cell height based on provided text.
func TestCalculateOrgInfoHeight(t *testing.T) {
	t.Parallel()
	options := pdf_service.Configuration()
	pdf := fpdf.New(options.PageLayout, options.PageUnits, options.PageDimensions, "")
	pdf.AddPage() // Ensure the page is added to calculate dimensions properly
	pdf.SetFont(options.FontFamily, "", options.FontSize)

	text := "Organization name that wraps over multiple lines"
	width := options.CardWidth
	cellHeight := options.CellHeight

	// Run the function and check for expected height calculation.
	calculatedHeight := pdf_service.CalculateOrgInfoHeight(pdf, text, width, cellHeight)
	expectedLines := 2
	expectedHeight := float64(expectedLines) * cellHeight

	assert.InDelta(t, expectedHeight, calculatedHeight, 10)
}

// TestCalculateCopyrightCellHeight verifies the calculation of cell height for copyright text.
func TestCalculateCopyrightCellHeight(t *testing.T) {
	t.Parallel()
	pdf := fpdf.New("P", "mm", "A4", "")
	options := pdf_service.Configuration()

	height := pdf_service.CalculateCopyrightCellHeight(pdf, options)
	expectedHeight := options.CellHeight * 0.4

	assert.InDelta(t, expectedHeight, height, 0.01)
}

// TestFindProductCodePairs verifies that product codes are paired correctly based on page height and margin.
func TestFindProductCodePairs(t *testing.T) {
	t.Parallel()
	options := pdf_service.Configuration()
	productCodeBoxes := map[string]float64{
		"Product1": 100.0,
		"Product2": 120.0,
		"Product3": 50.0,
		"Product4": 60.0,
		"Product5": 140.0,
	}

	pairs := pdf_service.FindProductCodePairs(productCodeBoxes, options)

	expectedPairs := [][2]string{
		{"Product5", "Product2"},
		{"Product1", "Product4"},
		{"Product3", ""},
	}

	assert.Equal(t, expectedPairs, pairs)
}
