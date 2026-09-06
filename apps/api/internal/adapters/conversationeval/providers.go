package conversationeval

import (
	"context"
	"sync/atomic"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type textProviders struct {
	providers  ports.RealtimeVoiceProviderSet
	explicit   ports.WorkflowLanguageProviderResolver
	transcript string
	calls      *atomic.Int64
}

func (p textProviders) ResolveRealtimeVoiceProviders(context.Context, ports.RealtimeVoiceProviderResolutionInput) (ports.RealtimeVoiceProviderSet, error) {
	result := p.providers
	result.SpeechToText = transcriptBridge{p.transcript}
	result.TextToSpeech = discardSpeech{}
	if result.ConversationModel != nil {
		result.ConversationModel = countedConversation{provider: result.ConversationModel, calls: p.calls}
	}

	return result, nil
}
func (p textProviders) ResolveWorkflowLanguageProvider(ctx context.Context, input ports.WorkflowLanguageProviderResolutionInput) (ports.WorkflowLanguageProviderBinding, error) {
	if p.explicit == nil {
		return ports.WorkflowLanguageProviderBinding{}, ErrInvalidExecution
	}
	resolved, err := p.explicit.ResolveWorkflowLanguageProvider(ctx, input)
	if err != nil {
		return ports.WorkflowLanguageProviderBinding{}, err
	}
	native := resolved.Provider
	if native == nil {
		return ports.WorkflowLanguageProviderBinding{}, ErrInvalidExecution
	}
	resolved.Provider = countedConversation{provider: native, calls: p.calls}
	return resolved, nil
}

type countedConversation struct {
	provider ports.ConversationModel
	calls    *atomic.Int64
}

func (p countedConversation) Converse(ctx context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	p.calls.Add(1)
	return p.provider.Converse(ctx, input)
}

// Text-only evaluation deliberately substitutes these two boundaries. No audio
// result is exposed and no speech-provider quality is claimed by this executor.
type transcriptBridge struct{ text string }

func (p transcriptBridge) Transcribe(ctx context.Context, _ ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	return ports.SpeechToTextResult{Transcript: p.text}, ctx.Err()
}

type discardSpeech struct{}

func (discardSpeech) Synthesize(ctx context.Context, _ ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	return ports.TextToSpeechResult{MimeType: "audio/mpeg", Chunks: [][]byte{{0}}}, ctx.Err()
}
