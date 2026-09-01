package backupmgr

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSaveAnalyzerCachesUnchangedSave(t *testing.T) {
	path := analysisFixture(t)
	analyzer := NewSaveAnalyzer(1, 2)
	var calls atomic.Int32
	analyzer.analyze = func(ctx context.Context, path string) (SaveAnalysis, error) {
		calls.Add(1)
		return AnalyzeSave(ctx, path)
	}

	first, err := analyzer.Analyze(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	second, err := analyzer.Analyze(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected one deep scan, got %d", calls.Load())
	}
	if first != second {
		t.Fatalf("cached result changed: first=%+v second=%+v", first, second)
	}
}

func TestSaveAnalyzerBoundsConcurrentScans(t *testing.T) {
	analyzer := NewSaveAnalyzer(1, 0)
	firstPath := analysisFixture(t)
	secondPath := analysisFixture(t)
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	var active atomic.Int32
	var maximum atomic.Int32
	analyzer.analyze = func(_ context.Context, _ string) (SaveAnalysis, error) {
		current := active.Add(1)
		for {
			old := maximum.Load()
			if current <= old || maximum.CompareAndSwap(old, current) {
				break
			}
		}
		started <- struct{}{}
		<-release
		active.Add(-1)
		return SaveAnalysis{}, nil
	}

	done := make(chan error, 2)
	go func() {
		_, err := analyzer.Analyze(context.Background(), firstPath)
		done <- err
	}()
	go func() {
		_, err := analyzer.Analyze(context.Background(), secondPath)
		done <- err
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first analysis did not start")
	}
	select {
	case <-started:
		t.Fatal("second analysis started before the concurrency slot was released")
	case <-time.After(25 * time.Millisecond):
	}
	close(release)
	for range 2 {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	if maximum.Load() != 1 {
		t.Fatalf("expected at most one concurrent scan, got %d", maximum.Load())
	}
}

func TestSaveAnalyzerWaitingRequestCanBeCancelled(t *testing.T) {
	analyzer := NewSaveAnalyzer(1, 0)
	path := analysisFixture(t)
	started := make(chan struct{})
	release := make(chan struct{})
	analyzer.analyze = func(_ context.Context, _ string) (SaveAnalysis, error) {
		close(started)
		<-release
		return SaveAnalysis{}, nil
	}

	firstDone := make(chan error, 1)
	go func() {
		_, err := analyzer.Analyze(context.Background(), path)
		firstDone <- err
	}()
	<-started

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := analyzer.Analyze(ctx, path); err != context.Canceled {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	close(release)
	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}
}
