package feed

import "tripmiles/dealsvc"

// LastKnown keeps the last counts for when the Deals Service does not answer.
type LastKnown interface {
	Save(snap dealsvc.Snapshot)
	AddBooking(dealID, partnerID string)
	Estimate() (dealsvc.Snapshot, bool)
}

// Creating interface here might be a little overcomplication
// but in production we might use this
