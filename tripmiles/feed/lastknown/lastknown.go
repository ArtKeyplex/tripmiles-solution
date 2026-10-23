// Package lastknown is a fallback for when the Deals Service does not answer, not a cache.
package lastknown

import (
	"maps"
	"sync"
	"time"

	"tripmiles/dealsvc"
)

const defaultMaxAge = 5 * time.Second

// Counts is the last answer plus this server's bookings since then.
type Counts struct {
	mu     sync.Mutex
	maxAge time.Duration
	snap   dealsvc.Snapshot // never changed after Save
	at     time.Time
	now    func() time.Time // nil means time.Now; tests set it

	dealBooked, partnerBooked map[string]int64
}

// New returns empty Counts.
func New() *Counts {
	// maxAge can be configured by taking it as a parameter of New.
	return &Counts{maxAge: defaultMaxAge}
}

// Save stores a fresh answer and drops the own bookings it already has.
func (c *Counts) Save(snap dealsvc.Snapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snap, c.at = snap, c.clock()
	c.dealBooked, c.partnerBooked = nil, nil
}

// AddBooking counts a booking of this server.
func (c *Counts) AddBooking(dealID, partnerID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dealBooked == nil {
		c.dealBooked, c.partnerBooked = map[string]int64{}, map[string]int64{}
	}
	c.dealBooked[dealID]++
	c.partnerBooked[partnerID]++
}

// Estimate returns the counts, or false if there is no answer younger than maxAge.
func (c *Counts) Estimate() (dealsvc.Snapshot, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.at.IsZero() || c.clock().Sub(c.at) > c.maxAge {
		return dealsvc.Snapshot{}, false
	}
	return c.withOwnBookings(), true
}

// withOwnBookings adds own bookings to a copy of the answer; the caller holds c.mu.
func (c *Counts) withOwnBookings() dealsvc.Snapshot {
	out := dealsvc.Snapshot{
		Deals:                maps.Clone(c.snap.Deals),
		PartnerBookingsToday: maps.Clone(c.snap.PartnerBookingsToday),
	}
	for id, n := range c.dealBooked {
		if d, ok := out.Deals[id]; ok {
			d.BookingsToday += n
			d.BookingsTotal += n
			out.Deals[id] = d
		}
	}
	for id, n := range c.partnerBooked {
		if _, ok := out.PartnerBookingsToday[id]; ok {
			out.PartnerBookingsToday[id] += n
		}
	}
	return out
}

func (c *Counts) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}
