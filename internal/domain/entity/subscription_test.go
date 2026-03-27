package entity

import "testing"

func TestCanTransitionTo_ValidTransitions(t *testing.T) {
	tests := []struct {
		name   string
		from   SubscriptionStatus
		to     SubscriptionStatus
		expect bool
	}{
		// From active
		{"active to past_due", SubscriptionActive, SubscriptionPastDue, true},
		{"active to cancelled", SubscriptionActive, SubscriptionCancelled, true},
		{"active to expired", SubscriptionActive, SubscriptionExpired, false},
		{"active to active", SubscriptionActive, SubscriptionActive, false},
		// From past_due
		{"past_due to active", SubscriptionPastDue, SubscriptionActive, true},
		{"past_due to cancelled", SubscriptionPastDue, SubscriptionCancelled, true},
		{"past_due to expired", SubscriptionPastDue, SubscriptionExpired, true},
		// From cancelled
		{"cancelled to expired", SubscriptionCancelled, SubscriptionExpired, true},
		{"cancelled to active", SubscriptionCancelled, SubscriptionActive, true},
		{"cancelled to past_due", SubscriptionCancelled, SubscriptionPastDue, false},
		// From expired
		{"expired to active", SubscriptionExpired, SubscriptionActive, true},
		{"expired to past_due", SubscriptionExpired, SubscriptionPastDue, false},
		{"expired to cancelled", SubscriptionExpired, SubscriptionCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)
			if got != tt.expect {
				t.Errorf("CanTransitionTo(%q -> %q) = %v, want %v", tt.from, tt.to, got, tt.expect)
			}
		})
	}
}

func TestValidTransitions_AllStatesHaveEntries(t *testing.T) {
	statuses := []SubscriptionStatus{
		SubscriptionActive,
		SubscriptionPastDue,
		SubscriptionCancelled,
		SubscriptionExpired,
	}
	for _, s := range statuses {
		if _, ok := ValidTransitions[s]; !ok {
			t.Errorf("ValidTransitions missing entry for status %q", s)
		}
	}
}
