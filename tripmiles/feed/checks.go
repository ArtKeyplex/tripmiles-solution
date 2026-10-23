package feed

import (
	"tripmiles/deals"
	"tripmiles/dealsvc"
)

// budgetChecks returns the deals that still have hotel stays left.
func (f *Feed) budgetChecks(ds []deals.Deal) []deals.Deal {
	snap := f.snapshot(ds)
	out := make([]deals.Deal, 0, len(ds))
	for _, d := range ds {
		if f.hasStaysLeft(d, snap) {
			out = append(out, d)
		}
	}
	return out
}

// hasStaysLeft checks all limits of d against all bookings, confirmed or not.
func (f *Feed) hasStaysLeft(d deals.Deal, snap dealsvc.Snapshot) bool {
	n, known := snap.Deals[d.ID]
	partnerBooked, partnerKnown := snap.PartnerBookingsToday[d.PartnerID]
	return fits(n.BookingsToday, known, d.DailyLimit) &&
		fits(n.BookingsTotal, known, d.TotalLimit) &&
		fits(partnerBooked, partnerKnown, f.Deals.Partner(d.PartnerID).DailyLimit)
}

func fits(booked int64, known bool, limit int64) bool {
	if limit == 0 {
		return true
	}
	return known && booked < limit
}

// snapshot asks the Deals Service once, without retries, and falls back to lastKnown.
func (f *Feed) snapshot(ds []deals.Deal) dealsvc.Snapshot {
	snap, err := f.DealsSvc.Snapshot(requestIDs(ds))
	if err != nil {
		if counts, ok := f.lastKnown.Estimate(); ok {
			return counts
		}
		// Unknown counts: deals with a limit get hidden.
		return dealsvc.Snapshot{}
	}
	// A parallel Book between the request and Save can be lost or counted twice.
	// Fix if needed: mark own bookings before the request and let Save drop only those.
	f.lastKnown.Save(snap)
	return snap
}

// requestIDs returns deal IDs and unique partner IDs.
func requestIDs(ds []deals.Deal) ([]string, []string) {
	dealIDs := make([]string, 0, len(ds))
	var partnerIDs []string
	seen := map[string]struct{}{}

	for _, d := range ds {
		dealIDs = append(dealIDs, d.ID)
		if _, ok := seen[d.PartnerID]; !ok {
			seen[d.PartnerID] = struct{}{}
			partnerIDs = append(partnerIDs, d.PartnerID)
		}
	}
	return dealIDs, partnerIDs
}
