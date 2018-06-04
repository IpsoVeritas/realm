package document

import (
	"sync"
	"time"
)

const MembershipApplicationType = "membership-application"

type MembershipApplication struct {
	Base
	AcceptedTerms string     `json:"acceptedTerms,omitempty"`
	Facts         *Multipart `json:"facts,omitempty"`
}

func NewMembershipApplication() *MembershipApplication {
	m := &MembershipApplication{
		Base: Base{
			Context:   Context,
			Type:      MembershipApplicationType,
			Timestamp: time.Now(),
			mu:        new(sync.Mutex),
		},
	}
	return m
}

const MembershipCancellationType = "membership-cancellation"

type MembershipCancellation struct {
	Base
	MembershipID string `json:"membershipId"`
	Reason       string `json:"reason,omitempty"`
}

func NewMembershipCancelation(membershipID, reason string) *MembershipCancellation {
	return &MembershipCancellation{
		Base: Base{
			Context:   Context,
			Type:      MembershipCancellationType,
			Timestamp: time.Now(),
			mu:        new(sync.Mutex),
		},
		MembershipID: membershipID,
		Reason:       reason,
	}
}
