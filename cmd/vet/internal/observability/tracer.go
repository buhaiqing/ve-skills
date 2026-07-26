package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type TraceContext struct {
	TraceID    string
	SpanID     string
	ParentID   string
	Service    string
	Operation  string
	StartTime  time.Time
	Attributes map[string]string
}

type Span struct {
	TraceContext
	Children []*Span
	Status   string
	Err      error
	mu       sync.Mutex
}

type traceKey struct{}

func NewRootTrace(service, operation string) *TraceContext {
	tc := &TraceContext{
		TraceID:    generateID(),
		SpanID:     generateID(),
		Service:    service,
		Operation:  operation,
		StartTime:  time.Now(),
		Attributes: make(map[string]string),
	}
	return tc
}

func StartSpan(parent *TraceContext, service, operation string) *Span {
	span := &Span{
		TraceContext: TraceContext{
			TraceID:    parent.TraceID,
			SpanID:     generateID(),
			ParentID:   parent.SpanID,
			Service:    service,
			Operation:  operation,
			StartTime:  time.Now(),
			Attributes: make(map[string]string),
		},
	}
	return span
}

func (s *Span) End(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		s.Status = "error"
		s.Err = err
	} else {
		s.Status = "ok"
	}
}

func (s *Span) Duration() time.Duration {
	return time.Since(s.StartTime)
}

func WithTrace(ctx context.Context, tc *TraceContext) context.Context {
	return context.WithValue(ctx, traceKey{}, tc)
}

func FromContext(ctx context.Context) *TraceContext {
	tc, _ := ctx.Value(traceKey{}).(*TraceContext)
	return tc
}

func generateID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}