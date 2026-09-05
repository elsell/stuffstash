package httpserver

func workflowDraftRequest() map[string]any {
	return map[string]any{
		"expectedRevision": 1,
		"definition": map[string]any{
			"name": "Home voice", "retrieval": "expanded", "response": "grounded",
			"budget": map[string]any{"evidenceRounds": 2, "modelCalls": 4, "elapsedSeconds": 30, "followUpTurns": 2},
			"steps": []map[string]any{
				{"kind": "interpret", "attempts": 1, "instructions": "Resolve existing items first."},
				{"kind": "assess", "attempts": 1},
				{"kind": "respond", "attempts": 1},
			},
		},
	}
}
