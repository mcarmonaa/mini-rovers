package rovers

import (
	"gopkg.in/src-d/go-queue.v1"
)

// Mention represents a message containing the repository required data.
type Mention struct {
	Endpoints []string
	IsFork    bool
}

// NewMention builds a new Mention.
func NewMention(endpoints []string, isFork bool) *Mention {
	return &Mention{
		Endpoints: endpoints,
		IsFork:    isFork,
	}
}

// PersistMentionFn is in charge to persist a Mention on any way.
type PersistMentionFn func(*Mention) error

// EnqueueMention generates a PersistMentionFn that send the mention to a queue.
func EnqueueMention(q queue.Queue) PersistMentionFn {
	return func(mention *Mention) error {
		job, err := queue.NewJob()
		if err != nil {
			return err
		}

		if err := job.Encode(mention); err != nil {
			return err
		}

		return q.Publish(job)
	}
}
