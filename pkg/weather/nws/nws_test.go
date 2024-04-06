package nws

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	metarData = map[string]string{
		"PAFA": "PAFA 061453Z 06003KT 10SM FEW070 SCT120 BKN140 M06/M09 A2962 RMK AO2 SLP038 T10611089 58008",
	}
	tafData = map[string]string{
		"PAFA": "PAFA 061120Z 0612/0718 04004KT P6SM FEW070 FM071300 05002KT P6SM VCSH OVC045",
	}
)

func TestUpdate(t *testing.T) {
	ts := createTestServer()
	defer ts.Close()

	wthr := NewNWSProviderOverride(
		fmt.Sprintf("%s/metar.xml.gz", ts.URL),
		fmt.Sprintf("%s/taf.xml.gz", ts.URL),
		ts.URL,
	)

	metars, tafs, err := wthr.Update()
	assert.NoError(t, err)
	assert.Equal(t, metarData, metars)
	assert.Equal(t, tafData, tafs)
}

func TestGetMETAR(t *testing.T) {
	ts := createTestServer()
	defer ts.Close()

	wthr := NewNWSProviderOverride(
		fmt.Sprintf("%s/metar.xml.gz", ts.URL),
		fmt.Sprintf("%s/taf.xml.gz", ts.URL),
		ts.URL,
	)

	obs, err := wthr.GetMETAR("PAFA")
	assert.NoError(t, err)
	assert.Equal(t, metarData["PAFA"], obs)
}

func TestGetTAF(t *testing.T) {
	ts := createTestServer()
	defer ts.Close()

	wthr := NewNWSProviderOverride(
		fmt.Sprintf("%s/metar.xml.gz", ts.URL),
		fmt.Sprintf("%s/taf.xml.gz", ts.URL),
		ts.URL,
	)

	fcst, err := wthr.GetTAF("PAFA")
	assert.NoError(t, err)
	assert.Equal(t, tafData["PAFA"], fcst)
}

func TestDownloadFile(t *testing.T) {
	ts := createTestServer()
	defer ts.Close()

	tests := []struct {
		name          string
		path          string
		expected      string
		statusCode    int
		expectedError bool
	}{
		{
			name:          "success",
			path:          "/ok",
			expected:      "ok",
			expectedError: false,
		},
		{
			name:          "bad request",
			path:          "/bad",
			expected:      "",
			expectedError: true,
		},
		{
			name:          "not found",
			path:          "/notfound",
			expected:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := downloadFile(ts.URL + tt.path)
			assert.Equal(t, tt.expected, strings.TrimSuffix(string(data), "\n"))
			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestProcessMETARs(t *testing.T) {
	expected := metarData

	ts := createTestServer()
	defer ts.Close()

	metars, err := processMETARs(ts.URL + "/metar.xml.gz")
	assert.NoError(t, err)
	assert.Equal(t, expected, metars)
}

func TestProcessTAFs(t *testing.T) {
	expected := tafData

	ts := createTestServer()
	defer ts.Close()

	tafs, err := processTAFs(ts.URL + "/taf.xml.gz")
	assert.NoError(t, err)
	assert.Equal(t, expected, tafs)
}

func createTestServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/metar.xml.gz":
			http.ServeFile(w, r, "testdata/metars.xml.gz")
		case "/taf.xml.gz":
			http.ServeFile(w, r, "testdata/tafs.xml.gz")
		case "/metar":
			http.ServeFile(w, r, "testdata/metars.json")
		case "/taf":
			http.ServeFile(w, r, "testdata/tafs.json")
		case "/bad":
			http.Error(w, "bad request", http.StatusBadRequest)
		case "/ok":
			fmt.Fprint(w, "ok")
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))

	return ts
}
