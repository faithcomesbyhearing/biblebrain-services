package copyright_test

import (
	"io"
	"testing"

	connection_service "biblebrain-services/service/connection"
	copyright_service "biblebrain-services/service/copyright"

	"github.com/stretchr/testify/require"
)

// TestStreamCopyrightIntegration verifies that the StreamCopyright method
// produces a valid PDF stream for a real database connection.
func TestStreamCopyrightIntegration(t *testing.T) {
	t.Parallel()
	sqlCon := connection_service.GetBibleBrainDB(t.Context())

	defer sqlCon.Close()

	mgr := copyright_service.New(sqlCon)

	// Define a package with known product codes
	pkg := copyright_service.Package{
		Products: []string{
			"P1PUI/LAN",
			"N2SWA/HNV",
			"N2POR/BSP",
			"N2ENG/NIV",
			"P1KEB/CIE",
		},
	}

	// Stream the PDF (audio=true)
	copyrights, err := mgr.GetCopyrightBy(t.Context(), pkg.Products, "audio")
	require.NoError(t, err)
	reader, err := mgr.StreamCopyright(t.Context(), copyrights, "audio")
	require.NoError(t, err)
	defer reader.Close()

	// Read the PDF header (should start with "%PDF-")
	header := make([]byte, 5)
	n, err := io.ReadFull(reader, header)
	require.NoError(t, err, "reading PDF header")
	require.Equal(t, 5, n)
	require.Equal(t, "%PDF-", string(header), "PDF header mismatch")

	// Read the rest into a buffer to ensure content flows
	rest, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NotEmpty(t, rest, "expected non-empty PDF content")
}

// TestGetCopyrightByIntegration verifies that GetCopyrightBy fetches copyright
// records from the database and includes valid organization info.
func TestGetCopyrightByIntegration(t *testing.T) {
	t.Parallel()
	sqlCon := connection_service.GetBibleBrainDB(t.Context())

	defer sqlCon.Close()

	mgr := copyright_service.New(sqlCon)

	codes := []string{"P1PUI/LAN", "N2SWA/HNV", "N2POR/BSP", "N2ENG/NIV", "P1KEB/CIE"}
	results, err := mgr.GetCopyrightBy(t.Context(), codes, "audio")
	require.NoError(t, err)
	require.NotEmpty(t, results, "expected at least one copyright record")

	// Check that each returned record has a valid ProductCode and Organizations
	for _, rec := range results {
		require.Contains(t, codes, rec.ProductCode, "unexpected product code")
		require.NotEmpty(t, rec.Organizations, "expected at least one organization")
	}
}
