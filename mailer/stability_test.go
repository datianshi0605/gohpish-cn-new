package mailer

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCanceledDialAndEmptyBatches(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	d := newMockDialer()
	sender, err := dialHost(ctx, d)
	if sender != nil || err != context.Canceled {
		t.Fatalf("expected cancellation, got %v %v", sender, err)
	}
	sendMail(ctx, d, generateMessages(d))
	sendMail(context.Background(), d, nil)
	mw := NewMailWorker()
	mw.Queue(nil)
	if d.dialCount != 0 {
		t.Fatal("canceled or empty batches must not dial")
	}
}

func TestConcurrentBatchesAreBounded(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mw := NewMailWorker()
	go mw.Start(ctx)
	release := make(chan struct{})
	started := make(chan struct{}, 32)
	var active, peak int32
	var finished sync.WaitGroup
	finished.Add(24)
	for i := 0; i < 24; i++ {
		d := newMockDialer()
		d.setDial(func() (Sender, error) {
			n := atomic.AddInt32(&active, 1)
			for {
				p := atomic.LoadInt32(&peak)
				if n <= p || atomic.CompareAndSwapInt32(&peak, p, n) {
					break
				}
			}
			started <- struct{}{}
			<-release
			atomic.AddInt32(&active, -1)
			finished.Done()
			return nil, context.Canceled
		})
		go mw.Queue(generateMessages(d))
	}
	for i := 0; i < maxConcurrentBatches; i++ {
		select {
		case <-started:
		case <-time.After(3 * time.Second):
			t.Fatal("batches failed to start")
		}
	}
	select {
	case <-started:
		t.Fatal("concurrency limit exceeded")
	case <-time.After(30 * time.Millisecond):
	}
	close(release)
	done := make(chan struct{})
	go func() { finished.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("waiting batches did not drain")
	}
	if atomic.LoadInt32(&peak) > maxConcurrentBatches {
		t.Fatal("too many simultaneous sends")
	}
}
