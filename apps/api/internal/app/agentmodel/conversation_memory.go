package agentmodel

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const DefaultConversationContextBytes = 2 * 1024 * 1024

var ErrConversationContextExhausted = errors.New("conversation context limit reached")

type ConversationScope struct {
	SessionID   string
	PrincipalID identity.PrincipalID
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
}

// ConversationMemory belongs to one active server-side session. Acquire leases
// its context until the caller commits or releases it; waiting is cancellable.
type ConversationMemory struct {
	scope     ConversationScope
	gate      chan struct{}
	messages  []ports.ConversationMessage
	maxBytes  int
	exhausted bool
}

func NewConversationMemory(scope ConversationScope, maxBytes int) *ConversationMemory {
	if maxBytes <= 0 {
		maxBytes = DefaultConversationContextBytes
	}
	return &ConversationMemory{scope: scope, gate: make(chan struct{}, 1), maxBytes: maxBytes}
}
func (m *ConversationMemory) Matches(scope ConversationScope) bool {
	return m != nil && m.scope == scope
}
func (m *ConversationMemory) Acquire(ctx context.Context, scope ConversationScope) ([]ports.ConversationMessage, error) {
	if !m.Matches(scope) {
		return nil, ports.ErrForbidden
	}
	select {
	case m.gate <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		m.Release()
		return nil, err
	}
	if m.exhausted {
		m.Release()
		return nil, ErrConversationContextExhausted
	}
	return m.messages, nil
}
func (m *ConversationMemory) Release() { <-m.gate }

// Commit is called while holding the lease. Exceeding the cap invalidates this
// context explicitly; a later utterance cannot run with silently truncated facts.
func (m *ConversationMemory) Commit(messages []ports.ConversationMessage) error {
	if err := checkConversationContext(messages, m.maxBytes); err != nil {
		m.messages = nil
		m.exhausted = true
		return ErrConversationContextExhausted
	}
	m.messages = messages
	return nil
}

func checkConversationContext(messages []ports.ConversationMessage, maxBytes int) error {
	if maxBytes <= 0 {
		maxBytes = DefaultConversationContextBytes
	}
	encoded, err := json.Marshal(messages)
	size := len(encoded)
	for _, message := range messages {
		size += len(message.ProviderState)
	}
	if err != nil || size > maxBytes {
		return ErrConversationContextExhausted
	}
	return nil
}
