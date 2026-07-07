package importjob

import (
	"testing"
	"time"
)

func TestAppendProgressHistoryRefreshesCollapsedPhase(t *testing.T) {
	t.Parallel()

	firstAt := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	secondAt := firstAt.Add(time.Minute)
	history := AppendProgressHistory(nil, Progress{
		Phase:     PhaseAssets,
		Done:      1,
		Total:     3,
		Message:   "Creating assets",
		UpdatedAt: firstAt,
	})

	history = AppendProgressHistory(history, Progress{
		Phase:     PhaseAssets,
		Done:      2,
		Total:     3,
		Message:   "Creating assets",
		UpdatedAt: secondAt,
	})

	if len(history) != 1 {
		t.Fatalf("expected repeated phase to stay collapsed, got %+v", history)
	}
	if history[0].Done != 2 || history[0].Total != 3 || !history[0].UpdatedAt.Equal(secondAt) {
		t.Fatalf("expected collapsed phase to refresh latest safe counts, got %+v", history[0])
	}
}

func TestAppendProgressHistoryKeepsDistinctMessagesInSamePhase(t *testing.T) {
	t.Parallel()

	firstAt := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	secondAt := firstAt.Add(time.Minute)
	history := AppendProgressHistory(nil, Progress{
		Phase:     PhaseAttachments,
		Done:      1,
		Total:     2,
		Message:   "Importing attachments",
		UpdatedAt: firstAt,
	})

	history = AppendProgressHistory(history, Progress{
		Phase:     PhaseAttachments,
		Done:      1,
		Total:     2,
		Message:   "Cancellation requested",
		UpdatedAt: secondAt,
	})

	if len(history) != 2 {
		t.Fatalf("expected same phase with a new safe message to remain distinct, got %+v", history)
	}
	if history[0].Message != "Importing attachments" || history[1].Message != "Cancellation requested" {
		t.Fatalf("expected distinct safe messages to be preserved, got %+v", history)
	}
}
