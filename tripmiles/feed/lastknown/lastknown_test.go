package lastknown

import (
	"testing"
	"time"

	"tripmiles/dealsvc"
)

// Here I'd prefer to use some library such as testify for tests, but it's better to use testing
// if it's already user across the project

func TestOldAnswerIsNotUsed(t *testing.T) {
	now := time.Now()
	c := New()
	c.now = func() time.Time { return now }
	c.Save(dealsvc.Snapshot{Deals: map[string]dealsvc.DealNumbers{"d": {}}})

	now = now.Add(defaultMaxAge)
	if _, ok := c.Estimate(); !ok {
		t.Fatal("The answer is exactly maxAge old. It was not returned.")
	}
	now = now.Add(time.Nanosecond)
	if _, ok := c.Estimate(); ok {
		t.Fatal("The answer is older than maxAge. It was still returned.")
	}
}

func TestNoAnswerBeforeTheFirstSave(t *testing.T) {
	c := New()
	if _, ok := c.Estimate(); ok {
		t.Fatal("Nothing was saved. Estimate returned an answer.")
	}
}
