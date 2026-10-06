package presentation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"sync"
	"testing"
	"time"
)

type failingReceiptOutput struct{ results, notices int }

func (o *failingReceiptOutput) Result(any) error    { o.results++; return errors.New("broken output") }
func (o *failingReceiptOutput) Notice(string) error { o.notices++; return errors.New("broken output") }
func TestReceiptOutputFailureDisablesOutputWithoutFailingOperation(t *testing.T) {
	output := &failingReceiptOutput{}
	sink := NewProtocolReceipts(output)
	receipt := ports.ProtocolReceipt{Operation: "claim", Result: ports.WorkerAttemptReceipt{}}
	sink.Record(context.Background(), receipt)
	sink.Record(context.Background(), receipt)
	sink.Close(context.Background())
	if output.results != 1 || output.notices != 1 {
		t.Fatalf("output repeated: %#v", output)
	}
}
func (*failingReceiptOutput) Error(string, string) {}

func TestConcurrentReceiptsRemainCompleteJSONRecords(t *testing.T) {
	var body bytes.Buffer
	sink := NewProtocolReceipts(Output{Stdout: &body, Stderr: &body, JSON: true})
	var group sync.WaitGroup
	for i := 0; i < 20; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			sink.Record(context.Background(), ports.ProtocolReceipt{Operation: "claim", Result: ports.WorkerAttemptReceipt{}})
		}()
	}
	group.Wait()
	sink.Close(context.Background())
	decoder := json.NewDecoder(&body)
	for i := 0; i < 20; i++ {
		var result struct {
			Operation string
			Result    struct{ Data *ports.ConsumerAttempt }
		}
		if err := decoder.Decode(&result); err != nil {
			t.Fatal(err)
		}
		if result.Operation != "claim" || result.Result.Data != nil {
			t.Fatalf("invalid record: %#v", result)
		}
	}
	if decoder.More() {
		t.Fatal("extra output")
	}
}

type blockedReceiptOutput struct {
	started chan struct{}
	release chan struct{}
}

func (o *blockedReceiptOutput) Result(any) error   { close(o.started); <-o.release; return nil }
func (*blockedReceiptOutput) Notice(string) error  { return nil }
func (*blockedReceiptOutput) Error(string, string) {}
func TestBlockedReceiptWriterCannotBlockWorkerOrCancellation(t *testing.T) {
	output := &blockedReceiptOutput{started: make(chan struct{}), release: make(chan struct{})}
	sink := NewProtocolReceipts(output)
	defer close(output.release)
	sink.Record(context.Background(), ports.ProtocolReceipt{Operation: "claim", Result: ports.WorkerAttemptReceipt{}})
	select {
	case <-output.started:
	case <-time.After(time.Second):
		t.Fatal("writer did not start")
	}
	returned := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			sink.Record(context.Background(), ports.ProtocolReceipt{})
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		sink.Close(ctx)
		close(returned)
	}()
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("receipt delivery blocked worker or cancellation")
	}
}
