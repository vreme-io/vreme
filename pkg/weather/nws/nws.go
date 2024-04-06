package nws

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type NWSProvider struct {
	addsMetarsCache string
	addsTafsCache   string
	apiRootURL      string
}

func NewNWSProvider() *NWSProvider {
	return NewNWSProviderOverride(
		"https://aviationweather.gov/data/cache/metars.cache.xml.gz",
		"https://aviationweather.gov/data/cache/tafs.cache.xml.gz",
		"https://aviationweather.gov/api/data",
	)
}

func NewNWSProviderOverride(metarCache, tafCache, apiRoot string) *NWSProvider {
	return &NWSProvider{
		addsMetarsCache: metarCache,
		addsTafsCache:   tafCache,
		apiRootURL:      apiRoot,
	}
}

// Update fetches the latest METARs and TAFs from the Aviation Weather Center
// and returns a map of METARs and TAFs.
func (n *NWSProvider) Update() (map[string]string, map[string]string, error) {
	metars, err := processMETARs(n.addsMetarsCache)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to process metars: %w", err)
	}

	tafs, err := processTAFs(n.addsTafsCache)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to process tafs: %w", err)
	}

	return metars, tafs, nil
}

func (n *NWSProvider) GetMETAR(station string) (string, error) {
	uri, err := n.getURL("metar", station)
	if err != nil {
		return "", fmt.Errorf("failed to get metar url: %w", err)
	}

	data, err := downloadFile(uri)
	if err != nil {
		return "", fmt.Errorf("failed to download METAR: %w", err)
	}

	metars := []dataAPIMETAR{}
	if err := json.Unmarshal(data, &metars); err != nil {
		return "", fmt.Errorf("failed to unmarshal METAR: %w", err)
	}

	return metars[0].Observation, nil
}

func (n *NWSProvider) GetTAF(station string) (string, error) {
	uri, err := n.getURL("taf", station)
	if err != nil {
		return "", fmt.Errorf("failed to get taf url: %w", err)
	}

	data, err := downloadFile(uri)
	if err != nil {
		return "", fmt.Errorf("failed to download TAF: %w", err)
	}

	tafs := []dataAPITAF{}
	if err := json.Unmarshal(data, &tafs); err != nil {
		return "", fmt.Errorf("failed to unmarshal TAF: %w", err)
	}

	return tafs[0].Forecast, nil
}

func (n *NWSProvider) getURL(endpoint, id string) (string, error) {
	u, err := url.Parse(
		fmt.Sprintf("%s/%s", n.apiRootURL, endpoint),
	)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}
	vals := u.Query()
	vals.Set("ids", id)
	vals.Set("format", "json")
	u.RawQuery = vals.Encode()
	return u.String(), nil
}

func processURL(cacheURL string) (*response, error) {
	metarsData, err := downloadFile(cacheURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download METARs: %w", err)
	}
	metarsData, err = ungzip(metarsData)
	if err != nil {
		return nil, fmt.Errorf("failed to ungzip METARs: %w", err)
	}
	return unmarshalResponse(metarsData)
}

func processTAFs(cacheURL string) (map[string]string, error) {
	xmlResponse, err := processURL(cacheURL)
	if err != nil {
		return nil, fmt.Errorf("failed to process tafs: %w", err)
	}

	tafs := make(map[string]string)
	for _, taf := range xmlResponse.TAFs {
		tafs[taf.StationID] = taf.RawText
	}

	return tafs, nil
}

func processMETARs(cacheURL string) (map[string]string, error) {
	xmlResponse, err := processURL(cacheURL)
	if err != nil {
		return nil, fmt.Errorf("failed to process metars: %w", err)
	}

	metars := make(map[string]string)
	for _, metar := range xmlResponse.METARs {
		metars[metar.StationID] = metar.RawText
	}

	return metars, nil
}

func unmarshalResponse(data []byte) (*response, error) {
	xmlResponse := &response{}
	if err := xml.Unmarshal(data, xmlResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return xmlResponse, nil
}

func downloadFile(path string) ([]byte, error) {
	client := &http.Client{
		Timeout: 2 * time.Minute,
	}

	resp, err := client.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func ungzip(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	return io.ReadAll(r)
}
