package filter

import (
	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/nozl/rate"
)

type (
	Filter interface {
		Allow() bool
		GetID() uuid.UUID
		SetFilterLimit(int)
		SetFilterBucketSize(int)
	}
	filter struct {
		id      uuid.UUID
		userID  string
		limiter *rate.Limiter
	}
)

func (fil *filter) Allow() bool {
	return fil.limiter.Allow()
}

func (fil *filter) GetID() uuid.UUID {
	return fil.id
}

func (fil *filter) SetFilterLimit(limit int) {
	fil.limiter.SetLimit(rate.Limit(limit))
}

func (fil *filter) SetFilterBucketSize(bucketSize int) {
	fil.limiter.SetBurst(bucketSize)
}

func NewFilter(limit int, bucket int, userID string) Filter {
	rl := rate.NewLimiter(rate.Limit(limit), bucket)

	return &filter{
		id:      uuid.New(),
		limiter: rl,
		userID:  userID,
	}
}
