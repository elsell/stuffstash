package voice

import "strings"

func googleGeminiBaseURL(cfg GoogleGeminiConfig) string {
	if strings.TrimSpace(cfg.BaseURL) != "" {
		return cfg.BaseURL
	}
	if strings.TrimSpace(cfg.APIKey) != "" {
		return "https://generativelanguage.googleapis.com"
	}
	return "https://" + googleGeminiLocation(cfg) + "-aiplatform.googleapis.com"
}

func googleGeminiPath(cfg GoogleGeminiConfig) string {
	if strings.TrimSpace(cfg.APIKey) != "" {
		return "/v1beta/" + googleGeminiAPIModelName(cfg.Model) + ":generateContent"
	}
	return "/v1/projects/" + cfg.ProjectID + "/locations/" + googleGeminiLocation(cfg) + "/publishers/google/models/" + cfg.Model + ":generateContent"
}

func googleGeminiAPIModelName(model string) string {
	model = strings.Trim(strings.TrimSpace(model), "/")
	if strings.HasPrefix(model, "models/") || strings.HasPrefix(model, "tunedModels/") {
		return model
	}
	return "models/" + model
}

func googleGeminiLocation(cfg GoogleGeminiConfig) string {
	location := strings.TrimSpace(cfg.Location)
	if location == "" {
		return "us-central1"
	}
	return location
}
