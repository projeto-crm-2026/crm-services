package entity

import "testing"

func TestPlanTableName(t *testing.T) {
	p := Plan{}
	if got := p.TableName(); got != "plans" {
		t.Errorf("Plan.TableName() = %q, want %q", got, "plans")
	}
}

func TestSubscriptionTableName(t *testing.T) {
	s := Subscription{}
	if got := s.TableName(); got != "subscriptions" {
		t.Errorf("Subscription.TableName() = %q, want %q", got, "subscriptions")
	}
}

func TestPlanLimits_FieldsExist(t *testing.T) {
	p := Plan{
		Name:              "test",
		DisplayName:       "Test",
		MaxContacts:       100,
		MaxMembers:        3,
		MaxChatResponders: 1,
	}
	if p.MaxContacts != 100 {
		t.Errorf("MaxContacts = %d, want 100", p.MaxContacts)
	}
	if p.MaxMembers != 3 {
		t.Errorf("MaxMembers = %d, want 3", p.MaxMembers)
	}
	if p.MaxChatResponders != 1 {
		t.Errorf("MaxChatResponders = %d, want 1", p.MaxChatResponders)
	}
}
