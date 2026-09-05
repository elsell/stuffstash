package bootstrap

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/adapters/conversationeval"
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/adapters/scheduling"
	"github.com/stuffstash/stuff-stash/internal/adapters/voice"
	modelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func buildEvaluationRuntime(cfg config.Config, settings config.EvaluationSettings, limits agentmodel.WorkflowLimits, observer ports.Observer, authorizer ports.Authorizer, repositories repositories, vault ports.ProviderCredentialVault) (modelapp.EvaluationRunCommandService, modelapp.EvaluationWorker) {
	clock := ports.SystemClock{}
	ids := idgen.NewULIDGenerator()
	resolver := voice.NewProviderProfileResolver(repositories.providerProfiles, repositories.voiceProviderConfigs, vault, googleProviderProfileFactory(cfg))
	executor := conversationeval.New(conversationeval.Dependencies{Clock: clock, IDs: ids, Observer: observer})
	commands := modelapp.NewEvaluationRunCommandService(modelapp.EvaluationRunCommandDependencies{Authorizer: authorizer, Runs: repositories.evaluationRuns, Workflows: repositories.conversationWorkflows, Cases: repositories.evaluationCases, Providers: resolver, IDs: ids, Clock: clock, Observer: observer, Limits: limits, MaxAttempts: settings.MaxAttempts})
	worker := modelapp.NewEvaluationWorker(modelapp.EvaluationWorkerDependencies{Runs: repositories.evaluationRuns, Authorizer: authorizer, Providers: resolver, Executor: executor, IDs: ids, Clock: clock, Observer: observer, LeaseGrace: settings.LeaseGrace, Delay: scheduling.Delay{}, PollInterval: settings.PollInterval})
	return commands, worker
}

type evaluationQueueDrainer interface {
	DrainEvaluationRuns(context.Context, int, int) error
}

func startEvaluationWorker(parent context.Context, application evaluationQueueDrainer, observer ports.Observer, cfg config.Config) (func(), error) {
	settings, err := cfg.ConversationEvaluations.Settings()
	if err != nil {
		return nil, err
	}
	if !settings.Enabled {
		return func() {}, nil
	}
	ctx, cancel := context.WithCancel(parent)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		runPeriodicDrain(ctx, settings.Interval, func() {
			if err := application.DrainEvaluationRuns(ctx, settings.DrainLimit, settings.Concurrency); err != nil && ctx.Err() == nil && observer != nil {
				observer.Record(ctx, ports.Event{Name: ports.EventConversationEvaluationQueueFailed, Message: "evaluation queue drain failed", Fields: map[string]string{"category": "processing_unavailable"}})
			}
		})
	}()
	return func() { cancel(); <-finished }, nil
}
