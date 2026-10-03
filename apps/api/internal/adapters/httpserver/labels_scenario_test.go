package httpserver

import (
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"testing"
)

func coverLabelScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	server, _, _ := labelTestServer(t)
	instance := coverage.request(t, server, "GET", "/instance", "/instance", "Bearer dev:other", nil, 200)
	if len(labelResponseData(t, instance.Body.Bytes())) != 2 {
		t.Fatal("public instance exposed more than protocol and identity")
	}
	created := performRequest(server, "POST", labelPrefix+"/assets", "Bearer dev:owner", map[string]any{"kind": "item", "title": "Tools"})
	if created.Code != 201 {
		t.Fatal(created.Body.String())
	}
	assetID := decodeAsset(t, created).Data.ID
	assetPath := labelPrefix + "/assets/" + assetID
	provisioned := performRequest(server, "POST", assetPath+"/label", "Bearer dev:owner", nil)
	if provisioned.Code != 200 {
		t.Fatal(provisioned.Body.String())
	}
	label := labelResponseData(t, provisioned.Body.Bytes())
	media := printing.MediaSnapshot{WidthMicrometers: 29000, HeightMicrometers: 90000, MarginsMicrometers: printing.Margins{Left: 1546, Right: 1546, Top: 3047, Bottom: 3048}, ResolutionDPI: 300, RasterWidth: 306, RasterHeight: 991, Orientation: printing.OrientationFeed, ColorMode: printing.ColorMonochrome, CutPolicy: printing.CutNone, DisplayRotation: 270}
	body := map[string]any{"media": media, "template": printing.TemplateSelection{ID: printing.TemplateQROnly, Version: 1}, "format": "png"}
	rendered := performRequest(server, "POST", assetPath+"/label-renders", "Bearer dev:owner", body)
	if rendered.Code != 201 {
		t.Fatal(rendered.Body.String())
	}
	contentPath := labelResponseData(t, rendered.Body.Bytes())["contentPath"].(string)
	token := "Bearer dev:owner"
	readStatus, renderStatus, resolveStatus := 200, 201, 200
	if adversarial {
		token = "Bearer dev:other"
		readStatus = 403
		renderStatus = 403
		resolveStatus = 404
	}
	const templateScope = "/tenants/{tenantId}/inventories/{inventoryId}"
	coverage.request(t, server, "POST", templateScope+"/assets/{assetId}/label", assetPath+"/label", token, nil, readStatus)
	coverage.request(t, server, "GET", templateScope+"/assets/{assetId}/label", assetPath+"/label", token, nil, readStatus)
	coverage.request(t, server, "POST", templateScope+"/assets/{assetId}/label-renders", assetPath+"/label-renders", token, body, renderStatus)
	coverage.request(t, server, "GET", templateScope+"/label-renders/{renderId}/content", contentPath, token, nil, readStatus)
	coverage.request(t, server, "GET", templateScope+"/label-templates", labelPrefix+"/label-templates", token, nil, readStatus)
	coverage.request(t, server, "GET", "/labels/v1/{instanceId}/{labelId}", "/labels/v1/"+label["instanceId"].(string)+"/"+label["labelId"].(string), token, nil, resolveStatus)
}
