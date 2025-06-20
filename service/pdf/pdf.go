package pdf

import (
	"sort"

	"github.com/go-pdf/fpdf"
)

type Options struct {
	FontSize       float64
	FontFamily     string
	AlignStrLeft   string
	FontStyle      string
	PageLayout     string
	PageUnits      string
	PageDimensions string

	PageHeight float64
	PageWidth  float64
	PageMargin float64

	CardsPerPage int

	CardWidth    float64
	CardPadding  float64
	BorderText   string
	CardHeight   float64
	ImgHeightMax float64
	ImgWidthMax  float64
	CellHeight   float64
}

func Configuration() Options {
	config := Options{
		PageDimensions: "A4", // A4, Letter, etc.
		PageLayout:     "P",  // L=Landscape, P=Portrait
		FontFamily:     "Arial",

		FontSize:  8,
		FontStyle: "", // Normal style
		PageUnits: "mm",

		CardsPerPage: 4, // "gridSize"
	}

	config.PageMargin = 8 // combined top and bottom margins

	if config.PageDimensions == "Letter" {
		config.PageWidth = 215.9
		config.PageHeight = 279.4
	} else { // assume A4
		config.PageWidth = 210.0
		config.PageHeight = 297.0
	}

	config.CardWidth = 99.25
	config.CardPadding = 1.0
	config.BorderText = "0"
	config.CardHeight = config.PageHeight/(float64(config.CardsPerPage)*0.5) - config.PageMargin

	config.ImgHeightMax = 20.0
	config.ImgWidthMax = 70.0
	config.CellHeight = 10

	if config.PageLayout == "L" {
		config.PageWidth, config.PageHeight = config.PageHeight, config.CardWidth
		config.CardWidth, config.CardHeight = config.CardHeight, config.CardWidth
	}

	return config
}

// Function to calculate the height needed for a cell based on the text and width.
func CalculateOrgInfoHeight(pdf *fpdf.Fpdf, text string, width, cellHeight float64) float64 {
	lines := pdf.SplitLines([]byte(text), width*0.81)

	return float64(len(lines)) * cellHeight
}

func CalculateCopyrightCellHeight(pdf *fpdf.Fpdf, o Options) float64 {
	pdf.SetFont(o.FontFamily, "", o.FontSize*0.90)

	return o.CellHeight * 0.4
}

// findProductCodePairs takes a map of product codes associated with their respective box heights
// and a Options struct containing PDF layout configurations. The function's primary goal is to
// pair up product codes in such a way that the combined height of each pair does not exceed a
// certain limit, which is determined by the pageHeight and pageMargin properties of the Options.
//
// The function returns a slice of string tuples, where each tuple contains either two paired product
// codes or a single unpaired code and an empty string.
func FindProductCodePairs(productCodeBoxes map[string]float64, opts Options) [][2]string {
	// Define a struct to hold key-value pairs.
	type keyValue struct {
		Key   string
		Value float64
	}

	// Create a slice to store key-value pairs extracted from the map.
	kvPairs := make([]keyValue, 0, len(productCodeBoxes))
	for k, v := range productCodeBoxes {
		kvPairs = append(kvPairs, keyValue{k, v})
	}

	// Sort the slice of key-value pairs in descending order based on the value (height).
	sort.Slice(kvPairs, func(i, j int) bool {
		return kvPairs[i].Value > kvPairs[j].Value
	})
	// Initialize a slice to hold unpaired product codes.
	var unpaired []string
	// Initialize a slice to store the pairs and a slice for unpaired product codes.
	pairs := make([][2]string, 0, (len(productCodeBoxes)+1)/2)
	// A map to keep track of which product codes have been used in pairs.
	used := make(map[string]bool)

	// Iterate over the sorted key-value pairs.
	for _, kvI := range kvPairs {
		codei, heighti := kvI.Key, kvI.Value

		// Skip this iteration if the product code has already been used.
		if used[codei] {
			continue
		}

		foundPair := false

		// Attempt to find a pair for the current product code.
		for _, kvJ := range kvPairs {
			codej, heightj := kvJ.Key, kvJ.Value
			// Check if a valid pair is found (different codes, not used, and combined height within limit).
			if codei != codej && !used[codej] && heighti+heightj <= opts.PageHeight-opts.PageMargin*4 {
				// Add the pair to the pairs slice and mark both codes as used.
				pairs = append(pairs, [2]string{codei, codej})
				used[codei] = true
				used[codej] = true
				foundPair = true

				break
			}
		}
		// If no pair is found, add the code to the unpaired slice.
		if !foundPair {
			unpaired = append(unpaired, codei)
			used[codei] = true
		}
	}
	// Append unpaired product codes at the end of the pairs slice, with an empty string as the second element.
	for _, code := range unpaired {
		pairs = append(pairs, [2]string{code, ""})
	}
	// Return the final slice of pairs.
	return pairs
}
