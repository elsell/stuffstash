package voice

import "github.com/stuffstash/stuff-stash/internal/ports"

func geminiActionPlanResponseSchema(tools []ports.AgentToolDescriptor) *geminiSchema {
	for _, tool := range tools {
		if tool.Name != "propose_action_plan" {
			continue
		}
		parameters := geminiParameters(tool.Parameters)
		return &geminiSchema{
			Type: "object",
			Properties: map[string]geminiSchema{
				"actionPlan": parameters,
			},
			Required: []string{"actionPlan"},
		}
	}
	return &geminiSchema{
		Type:       "object",
		Properties: map[string]geminiSchema{},
		Required:   []string{"actionPlan"},
	}
}
