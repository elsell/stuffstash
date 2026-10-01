package app

import (
	"encoding/json"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

// This catalog exposes the domain command contracts, not an intent taxonomy.
func realtimeConversationProposalTool() ports.ConversationToolDefinition {
	return ports.ConversationToolDefinition{
		Name:        realtimeConversationProposeTool,
		Description: "Prepare an inventory change for user approval; never execute it. Do not call this tool for an absolute date with an omitted year or ambiguous numeric notation such as 03/04: ask the user to clarify first. Relative dates such as next month are allowed after resolving them with get_expiration_calendar. Never infer a year for an incomplete absolute date or convert ambiguous day/month text to month precision. Preserve every explicit expiration date, and fetch the vocabulary manifest with {} before selecting its enabled customAssetTypeId. Search for existing items first. Use existing assetId/parentAssetId only from tool results. Commands may depend on earlier create commands via parentCommandId. Put all related commands in one ordered proposal; execution pauses for review immediately. Move existing items rather than duplicating them. An explicitly additional physical item may be created. Use a single update_asset command for name, description, custom field or expiration edits. Omit unchanged properties. Read the current detail and authorized vocabulary before field edits; clarify ambiguous labels. expiration object sets a date, null removes it. New field or type creation requires a separate single-command configuration proposal, inventory configure permission and explicit user approval. Use create_custom_field_definition or create_custom_asset_type only when the user intends to create schema. Never silently invent keys while editing an asset. Read current vocabulary first and clarify ambiguous field labels.",
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
                "type": "string"
              },
              "kind": {
                "type": "string",
                "enum": [
                  "update_asset"
                ]
              },
              "summary": {
                "type": "string"
              },
              "arguments": {
                "type": "object",
                "properties": {
                  "assetId": {
                    "type": "string"
                  },
                  "expiration": {
                    "anyOf": [
                      {
                        "type": "null"
                      },
                      {
                        "type": "object",
                        "properties": {
                          "date": {
                            "type": "string"
                          },
                          "precision": {
                            "type": "string",
                            "enum": [
                              "day",
                              "month"
                            ]
                          }
                        },
                        "required": [
                          "date",
                          "precision"
                        ],
                        "additionalProperties": false
                      }
                    ]
                  },
                  "customFields": {
                    "type": "object",
                    "minProperties": 1,
                    "maxProperties": 10,
                    "description": "Patch exact effective custom field keys from authorized vocabulary. Omit unchanged fields; null clears a known field. Clarify ambiguous field labels.",
                    "additionalProperties": {
                      "anyOf": [
                        {
                          "type": "string"
                        },
                        {
                          "type": "number"
                        },
                        {
                          "type": "boolean"
                        },
                        {
                          "type": "null"
                        }
                      ]
                    }
                  },
                  "title": {
                    "type": "string",
                    "minLength": 1
                  },
                  "description": {
                    "type": "string"
                  }
                },
                "required": [
                  "assetId"
                ],
                "additionalProperties": false,
                "minProperties": 2
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
                  },
                  "customAssetTypeId": {
                    "type": "string",
                    "description": "Existing assetTypeId from inventory vocabulary. Choose an expiration-enabled type when recording a date."
                  },
                  "expiration": {
                    "type": "object",
                    "properties": {
                      "date": {
                        "type": "string",
                        "description": "Exact YYYY-MM-DD or month-only YYYY-MM, preserving the user label. Clarify ambiguous or missing years."
                      },
                      "precision": {
                        "type": "string",
                        "enum": [
                          "day",
                          "month"
                        ]
                      }
                    },
                    "required": [
                      "date",
                      "precision"
                    ],
                    "additionalProperties": false
                  },
                  "customFields": {
                    "type": "object",
                    "minProperties": 1,
                    "maxProperties": 10,
                    "description": "Patch exact effective custom field keys from authorized vocabulary. Omit unchanged fields; null clears a known field. Clarify ambiguous field labels.",
                    "additionalProperties": {
                      "anyOf": [
                        {
                          "type": "string"
                        },
                        {
                          "type": "number"
                        },
                        {
                          "type": "boolean"
                        },
                        {
                          "type": "null"
                        }
                      ]
                    }
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
                  },
                  "customAssetTypeId": {
                    "type": "string",
                    "description": "Existing assetTypeId from inventory vocabulary. Choose an expiration-enabled type when recording a date."
                  },
                  "expiration": {
                    "type": "object",
                    "properties": {
                      "date": {
                        "type": "string",
                        "description": "Exact YYYY-MM-DD or month-only YYYY-MM, preserving the user label. Clarify ambiguous or missing years."
                      },
                      "precision": {
                        "type": "string",
                        "enum": [
                          "day",
                          "month"
                        ]
                      }
                    },
                    "required": [
                      "date",
                      "precision"
                    ],
                    "additionalProperties": false
                  },
                  "customFields": {
                    "type": "object",
                    "minProperties": 1,
                    "maxProperties": 10,
                    "description": "Patch exact effective custom field keys from authorized vocabulary. Omit unchanged fields; null clears a known field. Clarify ambiguous field labels.",
                    "additionalProperties": {
                      "anyOf": [
                        {
                          "type": "string"
                        },
                        {
                          "type": "number"
                        },
                        {
                          "type": "boolean"
                        },
                        {
                          "type": "null"
                        }
                      ]
                    }
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
          },
          {
            "type": "object",
            "properties": {
              "id": {
                "type": "string"
              },
              "kind": {
                "type": "string",
                "enum": [
                  "create_custom_asset_type"
                ]
              },
              "summary": {
                "type": "string"
              },
              "arguments": {
                "type": "object",
                "properties": {
                  "key": {
                    "type": "string"
                  },
                  "displayName": {
                    "type": "string"
                  },
                  "description": {
                    "type": "string"
                  },
                  "expirationEnabled": {
                    "type": "boolean"
                  }
                },
                "required": [
                  "key",
                  "displayName"
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
                "type": "string"
              },
              "kind": {
                "type": "string",
                "enum": [
                  "create_custom_field_definition"
                ]
              },
              "summary": {
                "type": "string"
              },
              "arguments": {
                "type": "object",
                "properties": {
                  "key": {
                    "type": "string"
                  },
                  "displayName": {
                    "type": "string"
                  },
                  "fieldType": {
                    "type": "string",
                    "enum": [
                      "text",
                      "number",
                      "boolean",
                      "date",
                      "url",
                      "enum"
                    ]
                  },
                  "applicability": {
                    "type": "string",
                    "enum": [
                      "all_assets",
                      "custom_asset_types"
                    ]
                  },
                  "enumOptions": {
                    "type": "array",
                    "maxItems": 50,
                    "items": {
                      "type": "string"
                    }
                  },
                  "customAssetTypeIds": {
                    "type": "array",
                    "maxItems": 10,
                    "items": {
                      "type": "string"
                    }
                  }
                },
                "required": [
                  "key",
                  "displayName",
                  "fieldType",
                  "applicability"
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
