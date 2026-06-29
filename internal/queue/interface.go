package queue

import "context"

type Queue interface {
	Publish(ctx context.Context, body []byte, attributes map[string]string) error
}
