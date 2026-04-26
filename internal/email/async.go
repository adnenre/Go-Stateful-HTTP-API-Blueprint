package email

import (
	"context"
	"sync"
)

type AsyncSender struct {
	inner Sender
	queue chan func()
	wg    sync.WaitGroup
}

func NewAsyncSender(inner Sender, workers int) *AsyncSender {
	as := &AsyncSender{
		inner: inner,
		queue: make(chan func(), 100),
	}
	for i := 0; i < workers; i++ {
		as.wg.Add(1)
		go as.worker()
	}
	return as
}

func (as *AsyncSender) worker() {
	defer as.wg.Done()
	for task := range as.queue {
		task()
	}
}

func (as *AsyncSender) SendOTP(ctx context.Context, to, otp string) error {
	as.queue <- func() {
		as.inner.SendOTP(ctx, to, otp)
	}
	return nil
}

func (as *AsyncSender) SendResetToken(ctx context.Context, to, token string) error {
	as.queue <- func() {
		as.inner.SendResetToken(ctx, to, token)
	}
	return nil
}

func (as *AsyncSender) Shutdown() {
	close(as.queue)
	as.wg.Wait()
}
