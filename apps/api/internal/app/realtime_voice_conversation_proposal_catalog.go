package app

import (
	"encoding/json"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

// This catalog exposes the domain command contracts, not an intent taxonomy.
func realtimeConversationProposalTool() ports.ConversationToolDefinition {
	return ports.ConversationToolDefinition{
		Name:        realtimeConversationProposeTool,
		Description: "Prepare an inventory change for user approval; never execute it. Search for existing items first. Use existing assetId/parentAssetId only from tool results. Commands may depend on earlier create commands via parentCommandId. Put all related commands in one ordered proposal; execution pauses for review immediately. Move existing items rather than duplicating them. An explicitly additional physical item may be created.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "summary": {
      "type": "string"
    },
    "risks": {
      "type": "array",
      "items": {
        "type": "string"
      }
    },
    "commands": {
      "type": "array",
      "minItems": 1,
      "maxItems": 10,
      "items": {
        "anyOf": [
          {
            "type": "object",
            "properties": {
              "id": {
                "type": "string",
                "description": "Unique plan-local command ID; this is not an asset ID."
              },
              "kind": {
                "type": "string",
                "enum": [
                  "create_asset"
                ]
              },
              "summary": {
                "type": "string",
                "description": "Concise proposed change for user review, without claiming it has executed."
              },
              "arguments": {
                "type": "object",
                "properties": {
                  "title": {
                    "type": "string"
                  },
                  "kind": {
                    "type": "string",
                    "enum": [
                      "item",
                      "container",
                      "location"
                    ]
                  },
                  "description": {
                    "type": "string"
                  },
                  "parentAssetId": {
                    "type": "string",
                    "description": "Existing parent ID from authorized tool results. Set at most one of parentAssetId and parentCommandId. Set the parent on a create directly; omit both parent fields for inventory root."
                  },
                  "parentCommandId": {
                    "type": "string",
                    "description": "ID of an earlier create command for the parent. Set at most one of parentAssetId and parentCommandId. Never use command IDs as assetId or parentAssetId."
                  }
                },
                "required": [
                  "title"
                ],
                "additionalProperties": false
              }
            },
            "required": [
              "id",
              "kind",
              "summary",
              "arguments"
            ],
            "additionalProperties": false
          },
          {
            "type": "object",
            "properties": {
              "id": {
                "type": "string",
                "description": "Unique plan-local command ID; this is not an asset ID."
              },
              "kind": {
                "type": "string",
                "enum": [
                  "create_location"
                ]
              },
              "summary": {
                "type": "string",
                "description": "Concise proposed change for user review, without claiming it has executed."
              },
              "arguments": {
                "type": "object",
                "properties": {
                  "title": {
                    "type": "string"
                  },
                  "kind": {
                    "type": "string",
                    "enum": [
                      "location"
                    ]
                  },
                  "description": {
                    "type": "string"
                  },
                  "parentAssetId": {
                    "type": "string",
                    "description": "Existing parent ID from authorized tool results. Set at most one of parentAssetId and parentCommandId. Set the parent on a create directly; omit both parent fields for inventory root."
                  },
                  "parentCommandId": {
                    "type": "string",
                    "description": "ID of an earlier create command for the parent. Set at most one of parentAssetId and parentCommandId. Never use command IDs as assetId or parentAssetId."
                  }
                },
                "required": [
                  "title"
                ],
                "additionalProperties": false
              }
            },
            "required": [
              "id",
              "kind",
              "summary",
              "arguments"
            ],
            "additionalProperties": false
          },
          {
            "type": "object",
            "properties": {
              "id": {
                "type": "string",
                "description": "Unique plan-local command ID; this is not an asset ID."
              },
              "kind": {
                "type": "string",
                "enum": [
                  "move_asset"
                ]
              },
              "summary": {
                "type": "string",
                "description": "Concise proposed change for user review, without claiming it has executed."
              },
              "arguments": {
                "type": "object",
                "properties": {
                  "assetId": {
                    "type": "string",
                    "description": "Existing asset ID returned by a tool. Cannot reference a create command or a not-yet-created asset."
                  },
                  "parentAssetId": {
                    "type": "string",
                    "description": "Existing parent ID from authorized tool results. Set at most one of parentAssetId and parentCommandId. Set the parent on a create directly; omit both parent fields for inventory root."
                  },
                  "parentCommandId": {
                    "type": "string",
                    "description": "ID of an earlier create command for the parent. Set at most one of parentAssetId and parentCommandId. Never use command IDs as assetId or parentAssetId."
                  }
                },
                "required": [
                  "assetId"
                ],
                "additionalProperties": false
              }
            },
            "required": [
              "id",
              "kind",
              "summary",
              "arguments"
            ],
            "additionalProperties": false
          },
          {
            "type": "object",
            "properties": {
              "id": {
                "type": "string",
                "description": "Unique plan-local command ID; this is not an asset ID."
              },
              "kind": {
                "type": "string",
                "enum": [
                  "archive_asset",
                  "restore_asset"
                ]
              },
              "summary": {
                "type": "string",
                "description": "Concise proposed change for user review, without claiming it has executed."
              },
              "arguments": {
                "type": "object",
                "properties": {
                  "assetId": {
                    "type": "string",
                    "description": "Existing asset ID returned by a tool. Cannot reference a create command or a not-yet-created asset."
                  }
                },
                "required": [
                  "assetId"
                ],
                "additionalProperties": false
              }
            },
            "required": [
              "id",
              "kind",
              "summary",
              "arguments"
            ],
            "additionalProperties": false
          },
          {
            "type": "object",
            "properties": {
              "id": {
                "type": "string",
                "description": "Unique plan-local command ID; this is not an asset ID."
              },
              "kind": {
                "type": "string",
                "enum": [
                  "checkout_asset",
                  "return_asset"
                ]
              },
              "summary": {
                "type": "string",
                "description": "Concise proposed change for user review, without claiming it has executed."
              },
              "arguments": {
                "type": "object",
                "properties": {
                  "assetId": {
                    "type": "string",
                    "description": "Existing asset ID returned by a tool. Cannot reference a create command or a not-yet-created asset."
                  },
                  "details": {
                    "type": "string"
                  }
                },
                "required": [
                  "assetId"
                ],
                "additionalProperties": false
              }
            },
            "required": [
              "id",
              "kind",
              "summary",
              "arguments"
            ],
            "additionalProperties": false
          }
        ]
      }
    }
  },
  "required": [
    "summary",
    "commands"
  ],
  "additionalProperties": false
}`),
	}
}
