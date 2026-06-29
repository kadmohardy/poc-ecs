package infra

import "sync"

type TraceRepository struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewTraceRepository() *TraceRepository {
	return &TraceRepository{
		data: make(map[string]string),
	}
}

func (r *TraceRepository) Save(endToEndID string, traceparent string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[endToEndID] = traceparent

	return nil
}

func (r *TraceRepository) Get(endToEndID string) (string, bool) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	traceparent, ok := r.data[endToEndID]

	return traceparent, ok
}

func (r *TraceRepository) Delete(endToEndID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.data, endToEndID)
}
