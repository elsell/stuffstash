package worklimit

import (
	"sync"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
)

func TestThumbnailReadinessScopesNotificationsAndForgetsSubscriptions(t *testing.T) {
	readiness := NewThumbnailReadiness()
	first, stopFirst := readiness.Watch(media.StorageKey("first"))
	defer stopFirst()
	second, stopSecond := readiness.Watch(media.StorageKey("first"))
	defer stopSecond()
	other, stopOther := readiness.Watch(media.StorageKey("other"))
	defer stopOther()
	cancelled, stopCancelled := readiness.Watch(media.StorageKey("first"))
	stopCancelled()
	stopCancelled()
	readiness.Published(media.StorageKey("first"))
	for _, ready := range []<-chan struct{}{first, second} {
		select {
		case <-ready:
		default:
			t.Fatal("published key did not wake reader")
		}
	}
	for _, ready := range []<-chan struct{}{other, cancelled} {
		select {
		case <-ready:
			t.Fatal("unrelated or cancelled reader was notified")
		default:
		}
	}
	readiness.Published(media.StorageKey("first"))
	fresh, stopFresh := readiness.Watch(media.StorageKey("first"))
	defer stopFresh()
	select {
	case <-fresh:
		t.Fatal("historical publication was retained")
	default:
	}
	stopFirst()
	readiness.Published(media.StorageKey("first"))
	select {
	case <-fresh:
	default:
		t.Fatal("old cancellation removed a newer subscription")
	}
}

func TestThumbnailReadinessConcurrentPublicationAndCancellation(t *testing.T) {
	readiness := NewThumbnailReadiness()
	var tasks sync.WaitGroup
	for range 100 {
		tasks.Add(1)
		go func() {
			defer tasks.Done()
			_, stop := readiness.Watch(media.StorageKey("shared"))
			readiness.Published(media.StorageKey("shared"))
			stop()
			stop()
		}()
	}
	tasks.Wait()
	ready, stop := readiness.Watch(media.StorageKey("shared"))
	defer stop()
	readiness.Published(media.StorageKey("shared"))
	select {
	case <-ready:
	default:
		t.Fatal("readiness stopped working after concurrent use")
	}
}
