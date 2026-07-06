package homebox

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (i LegacyImporter) withRequestOptions(request ports.ImportSourceRequest) LegacyImporter {
	client := *i.client
	if transport, ok := i.client.Transport.(*http.Transport); ok {
		cloned := transport.Clone()
		cloned.DialContext = guardedDialContext(request.AllowPrivateNetwork)
		if request.AllowInsecureTLS {
			cloned.TLSClientConfig = cloneTLSConfig(cloned.TLSClientConfig, true)
		}
		client.Transport = cloned
	} else if i.client.Transport == nil {
		cloned := http.DefaultTransport.(*http.Transport).Clone()
		cloned.DialContext = guardedDialContext(request.AllowPrivateNetwork)
		if request.AllowInsecureTLS {
			cloned.TLSClientConfig = cloneTLSConfig(cloned.TLSClientConfig, true)
		}
		client.Transport = cloned
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return ports.NewImportSourceUserError("too many Homebox redirects")
		}
		return validateOutboundURL(req.URL.String())
	}
	return LegacyImporter{client: &client, maxAttachmentBytes: i.maxAttachmentBytes, allowPrivateNetwork: request.AllowPrivateNetwork}
}

func guardedDialContext(allowPrivateNetwork bool) func(context.Context, string, string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 30 * time.Second}
	return func(ctx context.Context, network string, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		target, err := vettedDialAddress(ctx, host, port, allowPrivateNetwork)
		if err != nil {
			return nil, err
		}
		return dialer.DialContext(ctx, network, target)
	}
}

func vettedDialAddress(ctx context.Context, host string, port string, allowPrivateNetwork bool) (string, error) {
	if ip := net.ParseIP(host); ip != nil {
		if !allowPrivateNetwork && blockedOutboundIP(ip) {
			return "", ports.NewImportSourceUserError("Homebox URL resolves to a blocked address")
		}
		return net.JoinHostPort(ip.String(), port), nil
	}
	resolver := net.DefaultResolver
	ips, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return "", err
	}
	for _, candidate := range ips {
		if !allowPrivateNetwork && blockedOutboundIP(candidate.IP) {
			continue
		}
		return net.JoinHostPort(candidate.IP.String(), port), nil
	}
	return "", ports.NewImportSourceUserError("Homebox URL resolves to a blocked address")
}

func cloneTLSConfig(config *tls.Config, insecure bool) *tls.Config {
	if config == nil {
		config = &tls.Config{}
	} else {
		config = config.Clone()
	}
	config.InsecureSkipVerify = insecure
	return config
}

func safeLiveSourceError(err error) error {
	var userError ports.ImportSourceUserError
	if errors.As(err, &userError) {
		return userError
	}
	return err
}

func safeHomeboxWarningDetail(err error, fallback string) string {
	if err == nil {
		return fallback
	}
	return fallback
}

func validateOutboundURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return ports.NewImportSourceUserError("Homebox URL must use http or https")
	}
	host := parsed.Hostname()
	if strings.TrimSpace(host) == "" {
		return ports.NewImportSourceUserError("Homebox URL host is required")
	}
	return nil
}

func blockedOutboundIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified() ||
		ip.IsInterfaceLocalMulticast()
}

func normalizedMaxAttachmentBytes(value int64) int64 {
	if value <= 0 {
		return defaultMaxAttachmentBytes
	}
	return value
}

func homeboxFields() []importplan.FieldDefinition {
	return []importplan.FieldDefinition{
		{Key: "homebox-source-id", DisplayName: "Homebox Source ID", Type: "text"},
		{Key: "homebox-asset-id", DisplayName: "Homebox Asset ID", Type: "text"},
		{Key: "homebox-tags", DisplayName: "Homebox Tags", Type: "text"},
		{Key: "homebox-quantity", DisplayName: "Homebox Quantity", Type: "number"},
		{Key: "homebox-insured", DisplayName: "Homebox Insured", Type: "boolean"},
		{Key: "homebox-notes", DisplayName: "Homebox Notes", Type: "text"},
		{Key: "homebox-purchase-price", DisplayName: "Homebox Purchase Price", Type: "number"},
		{Key: "homebox-purchase-from", DisplayName: "Homebox Purchase From", Type: "text"},
		{Key: "homebox-purchase-time", DisplayName: "Homebox Purchase Time", Type: "text"},
		{Key: "homebox-manufacturer", DisplayName: "Homebox Manufacturer", Type: "text"},
		{Key: "homebox-model-number", DisplayName: "Homebox Model Number", Type: "text"},
		{Key: "homebox-serial-number", DisplayName: "Homebox Serial Number", Type: "text"},
		{Key: "homebox-lifetime-warranty", DisplayName: "Homebox Lifetime Warranty", Type: "boolean"},
		{Key: "homebox-warranty-expires", DisplayName: "Homebox Warranty Expires", Type: "text"},
		{Key: "homebox-warranty-details", DisplayName: "Homebox Warranty Details", Type: "text"},
		{Key: "homebox-sold-to", DisplayName: "Homebox Sold To", Type: "text"},
		{Key: "homebox-sold-price", DisplayName: "Homebox Sold Price", Type: "number"},
		{Key: "homebox-sold-time", DisplayName: "Homebox Sold Time", Type: "text"},
		{Key: "homebox-sold-notes", DisplayName: "Homebox Sold Notes", Type: "text"},
	}
}

func sourceReferenceFields(sourceID string) map[string]any {
	return map[string]any{"homebox-source-id": sourceID}
}

type legacyValues struct {
	AssetID          string
	Tags             string
	Quantity         string
	Insured          string
	Notes            string
	PurchasePrice    string
	PurchaseFrom     string
	PurchaseTime     string
	Manufacturer     string
	ModelNumber      string
	SerialNumber     string
	LifetimeWarranty string
	WarrantyExpires  string
	WarrantyDetails  string
	SoldTo           string
	SoldPrice        string
	SoldTime         string
	SoldNotes        string
}

func customFieldsFromLegacyValues(values legacyValues, sourceID string, sourceName string) (map[string]any, []importplan.Message) {
	fields := map[string]any{}
	var messages []importplan.Message
	addText := func(key string, value string) {
		value = strings.TrimSpace(value)
		if value != "" {
			fields[key] = value
		}
	}
	addNumber := func(key string, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		number, err := strconv.ParseFloat(value, 64)
		if err != nil {
			messages = append(messages, importplan.Message{Code: "invalid-number", Severity: importplan.SeverityWarning, Summary: "Homebox number could not be imported", Detail: key + "=" + value, SourceID: sourceID, SourceName: sourceName})
			return
		}
		fields[key] = number
	}
	addBool := func(key string, value string) {
		value = strings.TrimSpace(strings.ToLower(value))
		if value == "" {
			return
		}
		fields[key] = value == "true" || value == "1" || value == "yes"
	}
	addText("homebox-asset-id", values.AssetID)
	addText("homebox-source-id", sourceID)
	addText("homebox-tags", normalizedTags(values.Tags))
	addNumber("homebox-quantity", values.Quantity)
	addBool("homebox-insured", values.Insured)
	addText("homebox-notes", values.Notes)
	addNumber("homebox-purchase-price", values.PurchasePrice)
	addText("homebox-purchase-from", values.PurchaseFrom)
	addText("homebox-purchase-time", values.PurchaseTime)
	addText("homebox-manufacturer", values.Manufacturer)
	addText("homebox-model-number", values.ModelNumber)
	addText("homebox-serial-number", values.SerialNumber)
	addBool("homebox-lifetime-warranty", values.LifetimeWarranty)
	addText("homebox-warranty-expires", values.WarrantyExpires)
	addText("homebox-warranty-details", values.WarrantyDetails)
	addText("homebox-sold-to", values.SoldTo)
	addNumber("homebox-sold-price", values.SoldPrice)
	addText("homebox-sold-time", values.SoldTime)
	addText("homebox-sold-notes", values.SoldNotes)
	for _, value := range []string{values.PurchaseTime, values.WarrantyExpires, values.SoldTime} {
		if strings.HasPrefix(strings.TrimSpace(value), "0001-") {
			messages = append(messages, importplan.Message{
				Code:       "partial-date",
				Severity:   importplan.SeverityWarning,
				Summary:    "Homebox partial date imported as text",
				Detail:     value,
				SourceID:   sourceID,
				SourceName: sourceName,
			})
		}
	}
	return fields, messages
}

func (i LegacyImporter) login(ctx context.Context, baseURL string, username string, password string) (string, legacyStatus, error) {
	var loginResponse struct {
		Token string `json:"token"`
	}
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	if err := i.doJSON(ctx, http.MethodPost, baseURL+"/users/login", "", bytes.NewReader(body), &loginResponse); err != nil {
		return "", legacyStatus{}, err
	}
	var status legacyStatus
	if err := i.doJSON(ctx, http.MethodGet, baseURL+"/status", loginResponse.Token, nil, &status); err != nil {
		return "", legacyStatus{}, err
	}
	return loginResponse.Token, status, nil
}

func (i LegacyImporter) locations(ctx context.Context, baseURL string, token string) ([]legacyLocation, error) {
	var locations []legacyLocation
	return locations, i.doJSON(ctx, http.MethodGet, baseURL+"/locations", token, nil, &locations)
}

func (i LegacyImporter) locationTree(ctx context.Context, baseURL string, token string) ([]legacyTreeNode, error) {
	var tree []legacyTreeNode
	return tree, i.doJSON(ctx, http.MethodGet, baseURL+"/locations/tree", token, nil, &tree)
}

func (i LegacyImporter) items(ctx context.Context, baseURL string, token string) ([]legacyItemSummary, error) {
	var response struct {
		Items []legacyItemSummary `json:"items"`
	}
	if err := i.doJSON(ctx, http.MethodGet, baseURL+"/items", token, nil, &response); err != nil {
		return nil, err
	}
	return response.Items, nil
}

func (i LegacyImporter) item(ctx context.Context, baseURL string, token string, id string) (legacyItemDetail, error) {
	var detail legacyItemDetail
	return detail, i.doJSON(ctx, http.MethodGet, baseURL+"/items/"+url.PathEscape(id), token, nil, &detail)
}

func (i LegacyImporter) attachment(ctx context.Context, baseURL string, token string, itemID string, attachmentID string) ([]byte, string, error) {
	endpoint := baseURL + "/items/" + url.PathEscape(itemID) + "/attachments/" + url.PathEscape(attachmentID)
	if err := validateOutboundURL(endpoint); err != nil {
		return nil, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", token)
	resp, err := i.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("Homebox attachment returned %s", resp.Status)
	}
	maxBytes := normalizedMaxAttachmentBytes(i.maxAttachmentBytes)
	if resp.ContentLength > maxBytes {
		return nil, "", errors.New("Homebox attachment exceeds import size limit")
	}
	content, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, "", err
	}
	if int64(len(content)) > maxBytes {
		return nil, "", errors.New("Homebox attachment exceeds import size limit")
	}
	return content, sniffContentType(content, resp.Header.Get("Content-Type")), nil
}

func (i LegacyImporter) doJSON(ctx context.Context, method string, endpoint string, token string, body io.Reader, out any) error {
	if err := validateOutboundURL(endpoint); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ports.NewImportSourceUserError(fmt.Sprintf("Homebox returned %s", resp.Status))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

type legacyStatus struct {
	Build struct {
		Version string `json:"version"`
	} `json:"build"`
}

type legacyLocation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type legacyTreeNode struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Type     string           `json:"type"`
	Children []legacyTreeNode `json:"children"`
}

type legacyItemSummary struct {
	ID      string `json:"id"`
	AssetID string `json:"assetId"`
	Name    string `json:"name"`
}

type legacyItemDetail struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	Quantity         int                `json:"quantity"`
	Insured          bool               `json:"insured"`
	Archived         bool               `json:"archived"`
	PurchasePrice    float64            `json:"purchasePrice"`
	Location         legacyLocation     `json:"location"`
	Tags             []legacyTag        `json:"tags"`
	AssetID          string             `json:"assetId"`
	SerialNumber     string             `json:"serialNumber"`
	ModelNumber      string             `json:"modelNumber"`
	Manufacturer     string             `json:"manufacturer"`
	LifetimeWarranty bool               `json:"lifetimeWarranty"`
	WarrantyExpires  string             `json:"warrantyExpires"`
	WarrantyDetails  string             `json:"warrantyDetails"`
	PurchaseTime     string             `json:"purchaseTime"`
	PurchaseFrom     string             `json:"purchaseFrom"`
	SoldTime         string             `json:"soldTime"`
	SoldTo           string             `json:"soldTo"`
	SoldPrice        float64            `json:"soldPrice"`
	SoldNotes        string             `json:"soldNotes"`
	Notes            string             `json:"notes"`
	Attachments      []legacyAttachment `json:"attachments"`
}

type legacyTag struct {
	Name string `json:"name"`
}

type legacyAttachment struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Primary  bool   `json:"primary"`
	Title    string `json:"title"`
	MIMEType string `json:"mimeType"`
}

func legacyValuesFromItem(item legacyItemDetail) legacyValues {
	tags := make([]string, 0, len(item.Tags))
	for _, tag := range item.Tags {
		tags = append(tags, tag.Name)
	}
	return legacyValues{
		AssetID:          item.AssetID,
		Tags:             strings.Join(tags, "; "),
		Quantity:         strconv.Itoa(item.Quantity),
		Insured:          strconv.FormatBool(item.Insured),
		Notes:            item.Notes,
		PurchasePrice:    formatNumber(item.PurchasePrice),
		PurchaseFrom:     item.PurchaseFrom,
		PurchaseTime:     item.PurchaseTime,
		Manufacturer:     item.Manufacturer,
		ModelNumber:      item.ModelNumber,
		SerialNumber:     item.SerialNumber,
		LifetimeWarranty: strconv.FormatBool(item.LifetimeWarranty),
		WarrantyExpires:  item.WarrantyExpires,
		WarrantyDetails:  item.WarrantyDetails,
		SoldTo:           item.SoldTo,
		SoldPrice:        formatNumber(item.SoldPrice),
		SoldTime:         item.SoldTime,
		SoldNotes:        item.SoldNotes,
	}
}

func getCSV(row []string, header map[string]int, name string) string {
	index, ok := header[name]
	if !ok || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func normalizeBaseURL(value string) (string, error) {
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	if value == "" {
		return "", ports.NewImportSourceUserError("Homebox URL is required")
	}
	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		value = "https://" + value
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return "", ports.NewImportSourceUserError("Homebox URL is invalid")
	}
	if strings.HasSuffix(parsed.Path, "/api/v1") {
		return strings.TrimRight(value, "/"), nil
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/api/v1"
	return strings.TrimRight(parsed.String(), "/"), nil
}

func locationSourceID(value string) string {
	return "location:" + strings.TrimSpace(value)
}

func normalizedTags(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ';' || r == ','
	})
	tags := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		key := strings.ToLower(tag)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return strings.Join(tags, "; ")
}

func parseBool(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "true" || value == "1" || value == "yes"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func formatNumber(value float64) string {
	if value == 0 {
		return ""
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func imageImportLabel(include bool) string {
	if include {
		return "enabled"
	}
	return "disabled"
}

func sniffContentType(content []byte, fallback string) string {
	if len(content) > 0 {
		detected := http.DetectContentType(content)
		switch detected {
		case "image/jpeg", "image/png", "image/webp", "application/pdf":
			return detected
		}
	}
	if supportedImageType(fallback) {
		return strings.ToLower(strings.TrimSpace(fallback))
	}
	return strings.ToLower(strings.TrimSpace(fallback))
}

func supportedImageType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

func safeFileName(value string, fallback string) string {
	value = strings.TrimSpace(filepath.Base(value))
	if value == "." || value == "/" || value == "\\" || value == "" {
		return fallback
	}
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, "\\", "-")
	if len(value) > 255 {
		ext := filepath.Ext(value)
		stem := strings.TrimSuffix(value, ext)
		if len(ext) > 20 {
			ext = ""
		}
		if len(stem) > 255-len(ext) {
			stem = stem[:255-len(ext)]
		}
		value = stem + ext
	}
	return value
}

func defaultImageName(contentType string) string {
	switch contentType {
	case "image/png":
		return "homebox-image.png"
	case "image/webp":
		return "homebox-image.webp"
	default:
		return "homebox-image.jpg"
	}
}

func DecodeCSVBase64(value string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(strings.TrimSpace(value))
}
