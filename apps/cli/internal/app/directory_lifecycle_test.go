package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"testing"
)

type lifecycleAPI struct{ calls int }

func (a *lifecycleAPI) ChangeDirectoryLifecycle(context.Context, ports.DirectoryResource, ports.LifecycleAction, ports.Scope) (any, error) {
	a.calls++
	return map[string]string{"status": "deleted"}, nil
}

type lifecycleOutput struct{}

func (lifecycleOutput) Result(any) error     { return nil }
func (lifecycleOutput) Notice(string) error  { return nil }
func (lifecycleOutput) Error(string, string) {}

type lifecycleObserver struct{}

func (lifecycleObserver) Event(context.Context, string) {}
func TestLifecycleConfirmationPreventsUnapprovedWrites(t *testing.T) {
	api := &lifecycleAPI{}
	r := Runner{DirectoryLifecycle: func(string, string) (ports.DirectoryLifecycle, error) { return api, nil }, Output: lifecycleOutput{}, Observer: lifecycleObserver{}, Picker: firstScopeChoice{}}
	o := Options{Command: []string{"inventories", "delete"}, Scope: ports.Scope{Tenant: "home", Inventory: "garage"}}
	if err := r.directoryLifecycle(context.Background(), o, ports.Session{IDToken: "token"}); err == nil {
		t.Fatal("default cancel did not cancel")
	}
	for _, mode := range []Options{{Command: o.Command, Scope: o.Scope, JSON: true}, {Command: o.Command, Scope: o.Scope, NoInput: true}} {
		if err := r.directoryLifecycle(context.Background(), mode, ports.Session{IDToken: "token"}); err == nil {
			t.Fatal("script ran without --yes")
		}
	}
	if api.calls != 0 {
		t.Fatal("unapproved mutation sent")
	}
	o.Yes = true
	if err := r.directoryLifecycle(context.Background(), o, ports.Session{IDToken: "token"}); err != nil {
		t.Fatal(err)
	}
	if api.calls != 1 {
		t.Fatal("confirmed mutation not sent once")
	}
}

type failedCleanup struct{}

func (failedCleanup) Load(context.Context) (contexts.Config, error) { return contexts.Config{}, nil }
func (failedCleanup) Update(context.Context, func(*contexts.Config) error) error {
	return errors.New("disk unavailable")
}

type failedWarning struct {
	notices int
	result  bool
}

func (o *failedWarning) Notice(string) error {
	o.notices++
	if o.notices > 1 {
		return errors.New("stderr unavailable")
	}
	return nil
}
func (o *failedWarning) Result(any) error     { o.result = true; return nil }
func (o *failedWarning) Error(string, string) {}
func TestConfirmedDeleteSurvivesCleanupAndWarningFailure(t *testing.T) {
	api := &lifecycleAPI{}
	output := &failedWarning{}
	r := Runner{Contexts: failedCleanup{}, DirectoryLifecycle: func(string, string) (ports.DirectoryLifecycle, error) { return api, nil }, Output: output, Observer: lifecycleObserver{}}
	o := Options{Command: []string{"inventories", "delete"}, Scope: ports.Scope{Tenant: "home", Inventory: "garage"}, Yes: true}
	session := ports.Session{Issuer: "https://identity.example", Subject: "owner", IDToken: "token"}
	if err := r.directoryLifecycle(context.Background(), o, session); err != nil {
		t.Fatal(err)
	}
	if !output.result || api.calls != 1 {
		t.Fatal("confirmed result was suppressed")
	}
}
