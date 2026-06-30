package httpserver

import (
	"encoding/json"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func firstToolResultAssetID(content string, title string, lifecycleState string) (string, error) {
	ids, err := toolResultAssetIDs(content, lifecycleState)
	if err != nil {
		return "", err
	}
	if id := ids[title]; id != "" {
		return id, nil
	}
	return "", ports.ErrInvalidProviderInput
}

func firstToolResultAnyAssetID(content string, lifecycleState string, titles ...string) (string, error) {
	ids, err := toolResultAssetIDs(content, lifecycleState)
	if err != nil {
		return "", err
	}
	for _, title := range titles {
		if id := ids[title]; id != "" {
			return id, nil
		}
	}
	return "", ports.ErrInvalidProviderInput
}

func toolResultAssetIDs(content string, lifecycleState string) (map[string]string, error) {
	var output struct {
		Items []struct {
			AssetID        string `json:"assetId"`
			Title          string `json:"title"`
			LifecycleState string `json:"lifecycleState"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(content), &output); err != nil {
		return nil, ports.ErrInvalidProviderInput
	}
	ids := map[string]string{}
	for _, item := range output.Items {
		if item.LifecycleState == lifecycleState && item.AssetID != "" {
			ids[item.Title] = item.AssetID
		}
	}
	return ids, nil
}
