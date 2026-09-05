package httpserver

func evaluationCaseRequest() map[string]any {
	return map[string]any{"definition": map[string]any{"title": "Baby clothes", "utterance": "Where are my baby clothes?", "assets": []map[string]any{{"id": "box", "title": "Attic box", "kind": "container"}, {"id": "clothes", "title": "3 to 6 months", "kind": "item", "parentId": "box", "tagNames": []string{"baby", "clothes"}}}, "expectations": map[string]any{"kind": "answer", "referencedAssets": []string{"clothes"}, "locations": []map[string]string{{"assetId": "clothes", "ancestorId": "box"}}}}}
}
