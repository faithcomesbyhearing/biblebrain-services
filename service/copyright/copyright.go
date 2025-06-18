package copyright

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	pdf_service "biblebrain-services/service/pdf"
	sqlc "biblebrain-services/sqlc/generated"

	"github.com/go-pdf/fpdf"
)

const (
	SLASH      = "/"
	DASH       = "-"
	UNDERSCORE = "_"
	ModeAudio  = "audio"
	ModeVideo  = "video"
	ModeText   = "text"
)

type Package struct {
	Products []string `form:"productList"`
}

func (p Package) IsEmpty() bool {
	return len(p.Products) == 0
}

func (p Package) ID() string {
	var productCodes strings.Builder

	for i, product := range p.Products {
		if i > 0 {
			productCodes.WriteString(DASH)
		}
		// Replace SLASH with UNDERSCORE to avoid issues in file paths
		sanitizedProductCode := strings.ReplaceAll(product, SLASH, UNDERSCORE)
		productCodes.WriteString(sanitizedProductCode)
	}

	return productCodes.String()
}

type OrganizationsForCopyright struct {
	OrganizationID      uint   `json:"organizationId"`
	OrganizationSlug    string `json:"organizationSlug"`
	OrganizationName    string `json:"organizationName"`
	OrganizationLogoURL string `json:"organizationLogoUrl"`
}

type ByOrganizations struct {
	OrganizationIDList string `json:"-"`
	// it is an abstract struct to wrap the copyright information
	Organizations []OrganizationsForCopyright `json:"organizations"`
	// ProductCode is the product code for which the copyright applies.
	ProductCode   string `json:"productCode"`
	CopyrightDate string `json:"copyrightDate"`
	Copyright     string `json:"copyright"`
}

type LogoOrganization struct {
	URL    string
	Path   string
	Ext    string
	Width  float64
	Height float64
}

func (l LogoOrganization) HasValidPath() bool {
	return l.Path != ""
}

// Constants to define the structure and layout of the PDF.
const (
	LocalTempFolder        = "/tmp"
	CopyrightFolder        = "copyright"
	CopyrightGridAudio     = 4
	CopyrightGridVideo     = 8
	DirPerm                = 0o775
	MaxConcurrentDownloads = 8
)

type Service interface {
	GetCopyrightBy(ctx context.Context, productCodes []string, mode string) ([]ByOrganizations, error)
	StreamCopyright(ctx context.Context, copyrights []ByOrganizations, mode string) (io.ReadCloser, error)
}

// Define the struct that implements the interface.
type Manager struct {
	Connection *sql.DB
	Query      *sqlc.Queries
}

// Verify at compile-time that *CopyrightService implements Service.
var _ Service = (*Manager)(nil)

// Constructor for the struct.
func New(conn *sql.DB) *Manager {
	return &Manager{Connection: conn, Query: sqlc.New(conn)}
}

var ErrProductsNotFound = errors.New("no copyrights found for the provided product codes")

// StreamCopyright creates a PDF containing copyright information based on the provided package requests.
// If the package contains audio content, the layout is adjusted accordingly, otherwise, it's assumed to be video.
// The generated PDF is returned as an io.ReadCloser, allowing for streaming the PDF content directly.
//
// Parameters:
//
//	copyrights: A list of ByOrganizations detailing each copyright request.
//	isAudio: Boolean flag to determine if the package is audio (true) or video (false).
//
// Returns:
//
//	io.ReadCloser: A reader for the generated PDF content.
//	error: If there is any error during the process, otherwise nil.
func (m *Manager) StreamCopyright(
	ctx context.Context,
	copyrights []ByOrganizations,
	mode string,
) (io.ReadCloser, error) {
	if len(copyrights) == 0 {
		return nil, ErrProductsNotFound
	}

	var gridSize int

	if mode == ModeAudio {
		gridSize = CopyrightGridAudio
	} else {
		gridSize = CopyrightGridVideo
	}

	// Create a pipe:
	reader, writer := io.Pipe()

	go func() {
		// If ProducePdfCopyright fails, pipe EOF + error downstream.
		if err := ProducePdfCopyright(ctx, writer, copyrights, gridSize); err != nil {
			writer.CloseWithError(fmt.Errorf("generating PDF: %w", err))
		} else {
			writer.Close()
		}
	}()

	return reader, nil
}

// getTypeCodes returns a list of type codes based on the provided mode.
func getTypeCodes(mode string) []string {
	switch mode {
	case ModeAudio:
		return []string{"audio_drama", "audio"}
	case ModeVideo:
		return []string{"video_stream"}
	case ModeText:
		return []string{"text_plain", "text_html", "text_json", "text_format"}
	default:
		return []string{}
	}
}

// GetCopyrightBy retrieves copyright information for the specified product codes and mode.
func (m *Manager) GetCopyrightBy(
	ctx context.Context,
	productCodes []string,
	mode string,
) ([]ByOrganizations, error) {
	typeCodes := getTypeCodes(mode)

	// 1) Fetch the raw rows
	rows, err := m.Query.GetFilesetCopyrights(ctx, sqlc.GetFilesetCopyrightsParams{
		ProductCodes: productCodes,
		TypeCodes:    typeCodes,
	})
	if err != nil {
		slog.Error("fetching fileset copyrights", "error", err)

		return nil, fmt.Errorf("GetFilesetCopyrights: %w", err)
	}

	// 2) Parse & dedupe Organization IDs
	idSet := make(map[uint32]struct{}, len(rows))
	for _, r := range rows {
		for part := range strings.SplitSeq(r.OrganizationIDList.String, ",") {
			s := strings.TrimSpace(part)
			id64, err := strconv.ParseUint(s, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("parsing org ID %q: %w", s, err)
			}
			idSet[uint32(id64)] = struct{}{}
		}
	}

	// 3) Convert set to slice
	orgIDs := make([]uint32, 0, len(idSet))
	for id := range idSet {
		orgIDs = append(orgIDs, id)
	}

	// 4) Fetch organization details
	orgRows, err := m.Query.GetOrganizations(ctx, orgIDs)
	if err != nil {
		slog.Error("fetching organizations", "error", err)

		return nil, fmt.Errorf("GetOrganizations: %w", err)
	}

	// 5) Build a lookup map of orgID → OrganizationsForCopyright
	orgMap := make(map[uint]OrganizationsForCopyright, len(orgRows))
	for _, o := range orgRows {
		orgMap[uint(o.OrganizationID)] = OrganizationsForCopyright{
			OrganizationID:      uint(o.OrganizationID),
			OrganizationSlug:    o.OrganizationSlug,
			OrganizationName:    o.OrganizationName,
			OrganizationLogoURL: o.OrganizationLogoUrl.String,
		}
	}

	// 6) Assemble final slice, attaching the right orgs to each copyright
	out := make([]ByOrganizations, 0, len(rows))
	for _, row := range rows {
		entry := ByOrganizations{
			OrganizationIDList: row.OrganizationIDList.String,
			ProductCode:        row.ProductCode,
			CopyrightDate:      row.CopyrightDate.String,
			Copyright:          row.Copyright,
			Organizations:      []OrganizationsForCopyright{},
		}
		for part := range strings.SplitSeq(row.OrganizationIDList.String, ",") {
			s := strings.TrimSpace(part)
			id64, _ := strconv.ParseUint(s, 10, 32)
			if org, ok := orgMap[uint(id64)]; ok {
				entry.Organizations = append(entry.Organizations, org)
			}
		}
		out = append(out, entry)
	}

	return out, nil
}

// producePdfCopyright generates a PDF file consisting of copyright information.
// The generated PDF has a predefined layout with copyright entries placed in a grid structure.
// Each copyright entry includes an organization's logo and relevant details.
//
// Parameters:
//
//	copyrights: A list of ByOrganizations detailing each organization's copyright.
//	targetPdfFile: The desired path for the resulting PDF file.
//	gridSize: Specifies the grid size, which determines the number of copyright entries per page.
//
// Returns:
//
//	error: If there is any error during the PDF generation process, otherwise nil.
func ProducePdfCopyright(
	ctx context.Context,
	writer io.Writer,
	copyrights []ByOrganizations,
	gridSize int,
) error {
	opts := pdf_service.Configuration()

	pdf := fpdf.New(opts.PageLayout, opts.PageUnits, opts.PageDimensions, "")
	pdf.SetTitle("Copyright", false)
	pdf.SetAuthor("Biblebrain Downloader", false)
	pdf.SetFont(opts.FontFamily, opts.FontStyle, opts.FontSize)

	const cardsPerRow = 2
	padding := opts.PageMargin / cardsPerRow

	var axisY float64
	var axisX float64

	heightByCopyright := make(map[string]float64)
	copyrightPeerProdCode := make(map[string]ByOrganizations)

	downloadedImages := downloadOrgLogos(ctx, copyrights)
	placedCards := 0
	placedTuples := 0
	logos := make(map[string]LogoOrganization)

	for _, copyright := range copyrights {
		copyrightPeerProdCode[copyright.ProductCode] = copyright

		for _, org := range copyright.Organizations {
			logo, exists := downloadedImages[org.OrganizationLogoURL]
			if !exists {
				continue
			}

			imageBaseName := path.Base(logo)
			imageExt := path.Ext(imageBaseName)

			newPNGImage := logo
			var imageWidth float64
			var imageHeight float64

			if imageExt == "."+SVGFormat {
				newPNGImage, _ = SVGToPNG(logo)
				imageExt = PNGFormat
			}

			widthNewPNGImage, heightNewPNGImage, err := GetImageDimensions(newPNGImage)
			if err != nil {
				slog.Error("Failed to get image dimensions", "error", err.Error(), "image", newPNGImage)

				continue
			}
			imageWidth, imageHeight = ScaleDimensions(widthNewPNGImage, heightNewPNGImage, opts.ImgWidthMax, opts.ImgHeightMax)
			currentImageHeight := math.Min(imageHeight, opts.ImgHeightMax)
			logos[org.OrganizationLogoURL] = LogoOrganization{
				URL:    org.OrganizationLogoURL,
				Path:   newPNGImage,
				Ext:    imageExt,
				Width:  imageWidth,
				Height: currentImageHeight,
			}
		}

		heightByCopyright[copyright.ProductCode] = 6 // header
		const threshold = 0.5                        // Threshold to avoid too small cards
		for _, org := range copyright.Organizations {
			if logoOrganization, ok := logos[org.OrganizationLogoURL]; ok {
				heightByCopyright[copyright.ProductCode] += logoOrganization.Height
			}
			heightByCopyright[copyright.ProductCode] += pdf_service.CalculateOrgInfoHeight(
				pdf,
				org.OrganizationName,
				opts.CardWidth,
				opts.CellHeight,
			) * threshold
			heightByCopyright[copyright.ProductCode] += opts.CellHeight * threshold // Copyright Date
			heightByCopyright[copyright.ProductCode] += cardsPerRow
		}

		cellHeightCopyRight := pdf_service.CalculateCopyrightCellHeight(pdf, opts)
		copyRightLines := pdf.SplitLines([]byte(copyright.Copyright), opts.CardWidth-opts.CardPadding*2)
		heightByCopyright[copyright.ProductCode] += float64(len(copyRightLines)) * cellHeightCopyRight
	}

	productCodePairs := pdf_service.FindProductCodePairs(heightByCopyright, opts)

	for _, codeTuple := range productCodePairs {
		if placedTuples%gridSize == 0 {
			pdf.AddPage()
			axisY = padding // Reset Y to top of the page
		}

		axisX = padding + float64(placedCards%cardsPerRow)*(opts.CardWidth+padding)

		var code1 string
		var code2 string
		var firstCardHeight float64
		var secondCardHeight float64

		if code1 = codeTuple[0]; code1 != "" {
			remainingHeight := opts.CardHeight - heightByCopyright[code1]
			firstCardHeight = heightByCopyright[code1]

			if remainingHeight > 0 {
				firstCardHeight = heightByCopyright[code1] + remainingHeight
			}
			placeCard(pdf, opts, copyrightPeerProdCode[code1], logos, axisX, axisY, opts.CardWidth, firstCardHeight)
		}
		if code2 = codeTuple[1]; code2 != "" {
			remainingHeight := opts.CardHeight*cardsPerRow - firstCardHeight
			secondCardHeight = max(remainingHeight, heightByCopyright[code2])
			placeCard(
				pdf,
				opts,
				copyrightPeerProdCode[code2],
				logos,
				axisX,
				firstCardHeight+padding*cardsPerRow,
				opts.CardWidth,
				secondCardHeight,
			)
		}

		placedCards++
		placedTuples += cardsPerRow
	}

	if err := pdf.Output(writer); err != nil {
		return fmt.Errorf("writing PDF to writer: %w", err)
	}

	return nil
}

// downloadOrgLogos downloads logos of organizations.
// It takes in a slice of copyrights, each containing information about an organization
// including its logo URL. The function returns a map where the keys are logo URLs and the
// values are the file paths where the logos were downloaded to.
func downloadOrgLogos(ctx context.Context, copyrights []ByOrganizations) map[string]string {
	type result struct {
		url, path string
		err       error
	}
	tmpDir := filepath.Join(LocalTempFolder, CopyrightFolder)
	// Ensure the base tmpDir exists
	if err := os.MkdirAll(tmpDir, DirPerm); err != nil {
		slog.Error("failed to create tmpDir", "path", tmpDir, "err", err)

		return nil
	}

	// Collect all distinct URLs
	urlSet := make(map[string]struct{})

	for _, cr := range copyrights {
		for _, org := range cr.Organizations {
			urlSet[org.OrganizationLogoURL] = struct{}{}
		}
	}
	urls := make([]string, 0, len(urlSet))

	for url := range urlSet {
		urls = append(urls, url)
	}

	channel := make(chan result, len(urls))
	sem := make(chan struct{}, MaxConcurrentDownloads)

	for _, url := range urls {
		sem <- struct{}{}
		go func(url string) {
			defer func() { <-sem }()

			dest := filepath.Join(tmpDir, path.Base(url))

			err := DownloadImage(ctx, url, dest)

			if err != nil {
				slog.Warn("download failed", "url", url, "err", err)
				channel <- result{url, "", err}
			} else {
				channel <- result{url, dest, nil}
			}
		}(url)
	}

	downloaded := make(map[string]string, len(urls))
	// Wait for all goroutines
	for range urls {
		res := <-channel
		if res.err == nil {
			downloaded[res.url] = res.path
		}
	}

	return downloaded
}

func placeCard(pdf *fpdf.Fpdf, opts pdf_service.Options,
	copyright ByOrganizations,
	pathOrgLogo map[string]LogoOrganization,
	axisX float64, axisY float64,
	cardWidth float64,
	cardHeight float64,
) {
	const cardPadding = 2
	currentY := axisY

	// Draw Rectangle for current card
	pdf.Rect(axisX, currentY, opts.CardWidth, cardHeight, "D")
	currentY += cardPadding

	// Place product code as title for current card
	productCode := copyright.ProductCode
	midpointCard := (cardWidth / cardPadding)
	midpointTitle := pdf.GetStringWidth(productCode) / cardPadding
	startLocation := axisX + midpointCard - midpointTitle

	pdf.SetFont(opts.FontFamily, "", opts.FontSize)
	pdf.SetXY(startLocation, currentY)
	pdf.Write(cardPadding, productCode)

	currentY += cardPadding

	// Draw Organization information (Logo, name, etc.)
	for _, org := range copyright.Organizations {
		orgInfoHeight := placeOrgInfo(pdf, opts, copyright, org, pathOrgLogo[org.OrganizationLogoURL], axisX, currentY)
		currentY += orgInfoHeight
	}

	pdf.SetXY(axisX+opts.CardPadding, currentY)
	tr := pdf.UnicodeTranslatorFromDescriptor("") // Fixes rendering of unicode
	pdf.MultiCell(
		opts.CardWidth-opts.CardPadding*cardPadding,
		pdf_service.CalculateCopyrightCellHeight(pdf, opts),
		tr(copyright.Copyright),
		opts.BorderText,
		"",
		false,
	)
}

func placeOrgInfo(pdf *fpdf.Fpdf, opts pdf_service.Options,
	copyright ByOrganizations,
	copyrightOrg OrganizationsForCopyright,
	orgLogo LogoOrganization,
	axisX float64, axisY float64,
) float64 {
	var opt fpdf.ImageOptions
	opt.ReadDpi = true

	opt.ImageType = strings.TrimPrefix(orgLogo.Ext, ".")

	currentY := axisY

	if orgLogo.HasValidPath() {
		pdf.ImageOptions(orgLogo.Path, axisX+opts.CardPadding, currentY, orgLogo.Width, orgLogo.Height, false, opt, 0, "")
	}

	currentY += orgLogo.Height

	organizationName := copyrightOrg.OrganizationName
	copyrightDate := copyright.CopyrightDate

	headerOrgNameHeight := pdf_service.CalculateOrgInfoHeight(pdf, organizationName, opts.CardWidth, opts.CellHeight)

	orgNameLabel := "Org. Name: "
	dateLabel := "Copyright Date: "

	pdf.SetFontStyle("B") // Bold

	const headerHeightFactor = 0.5
	const paddingHeight = 2
	// Labels
	pdf.SetXY(axisX+opts.CardPadding, currentY)
	pdf.MultiCell(
		opts.CardWidth*0.19,
		headerOrgNameHeight*headerHeightFactor,
		orgNameLabel,
		opts.BorderText,
		opts.AlignStrLeft,
		false)
	pdf.SetXY(axisX+opts.CardPadding, currentY+headerOrgNameHeight*headerHeightFactor)
	pdf.MultiCell(
		opts.CardWidth*0.25,
		opts.CellHeight*headerHeightFactor,
		dateLabel,
		opts.BorderText,
		opts.AlignStrLeft,
		false,
	)

	// Values
	pdf.SetFontStyle("") // Normal
	pdf.SetXY(axisX+opts.CardPadding+(opts.CardWidth*0.19), currentY)
	pdf.MultiCell(
		opts.CardWidth*0.81-opts.CardPadding,
		opts.CellHeight*headerHeightFactor,
		organizationName,
		opts.BorderText,
		opts.AlignStrLeft,
		false,
	)

	pdf.SetXY(axisX+opts.CardPadding+(opts.CardWidth*0.25), currentY+headerOrgNameHeight*headerHeightFactor)
	pdf.MultiCell(
		opts.CardWidth*0.13,
		opts.CellHeight*headerHeightFactor,
		copyrightDate,
		opts.BorderText,
		opts.AlignStrLeft,
		false,
	)

	// Account for height of drawn text
	currentY += opts.CellHeight*headerHeightFactor + headerOrgNameHeight*headerHeightFactor
	currentY += paddingHeight // Add padding after the header

	// restore font settings
	pdf.SetFont(opts.FontFamily, opts.FontStyle, opts.FontSize)

	orgInfoHeight := currentY - axisY

	return orgInfoHeight
}
