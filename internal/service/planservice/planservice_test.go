package planservice

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type mockPlanRepo struct {
	plans          []entity.Plan
	planByID       *entity.Plan
	planByName     *entity.Plan
	contactCount   int
	memberCount    int
	responderCount int
	err            error
}

func (m *mockPlanRepo) GetAll(ctx context.Context) ([]entity.Plan, error) {
	return m.plans, m.err
}
func (m *mockPlanRepo) GetByID(ctx context.Context, id uint) (*entity.Plan, error) {
	return m.planByID, m.err
}
func (m *mockPlanRepo) GetByName(ctx context.Context, name string) (*entity.Plan, error) {
	return m.planByName, m.err
}
func (m *mockPlanRepo) CountContacts(ctx context.Context, orgID uuid.UUID) (int, error) {
	return m.contactCount, m.err
}
func (m *mockPlanRepo) CountMembers(ctx context.Context, orgID uuid.UUID) (int, error) {
	return m.memberCount, m.err
}
func (m *mockPlanRepo) CountChatResponders(ctx context.Context, orgID uuid.UUID) (int, error) {
	return m.responderCount, m.err
}
func (m *mockPlanRepo) GetByMpPlanID(ctx context.Context, mpPlanID string) (*entity.Plan, error) {
	return m.planByID, m.err
}
func (m *mockPlanRepo) UpdateMpPlanID(ctx context.Context, planID uint, mpPlanID string) error {
	return m.err
}

type mockSubRepo struct {
	sub *entity.Subscription
	err error
}

func (m *mockSubRepo) Create(ctx context.Context, sub *entity.Subscription) (*entity.Subscription, error) {
	return sub, m.err
}
func (m *mockSubRepo) GetByOrganizationID(ctx context.Context, orgID uuid.UUID) (*entity.Subscription, error) {
	return m.sub, m.err
}
func (m *mockSubRepo) GetByID(ctx context.Context, id uint) (*entity.Subscription, error) {
	return m.sub, m.err
}
func (m *mockSubRepo) UpdateStatus(ctx context.Context, id uint, status entity.SubscriptionStatus) error {
	return m.err
}
func (m *mockSubRepo) Update(ctx context.Context, sub *entity.Subscription) error {
	return m.err
}
func (m *mockSubRepo) GetByMpSubscriptionID(ctx context.Context, mpSubID string) (*entity.Subscription, error) {
	return m.sub, m.err
}

type mockUserRepo struct{}

func (m *mockUserRepo) Insert(ctx context.Context, name, email, passwordHash string, organizationID uuid.UUID) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByInviteToken(ctx context.Context, token string) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) InsertPending(ctx context.Context, name, email, inviteToken string, inviteExpiry time.Time, organizationID uuid.UUID, invitedBy uint) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) ActivateUser(ctx context.Context, userID uint, passwordHash string) error {
	return nil
}
func (m *mockUserRepo) UpdateRoleID(ctx context.Context, userID, roleID uint) error { return nil }
func (m *mockUserRepo) DeactivateUser(ctx context.Context, userID uint) error      { return nil }
func (m *mockUserRepo) ReactivateUser(ctx context.Context, userID uint) error       { return nil }
func (m *mockUserRepo) RemoveMember(ctx context.Context, userID uint) error         { return nil }
func (m *mockUserRepo) ListByOrganization(ctx context.Context, organizationID uuid.UUID) ([]entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) ListByOrganizationWithRole(ctx context.Context, organizationID uuid.UUID) ([]entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetOrgOwnerEmail(ctx context.Context, orgID uuid.UUID) (string, string, error) {
	return "Owner", "owner@test.com", nil
}

type mockMailer struct{}

func (m *mockMailer) SendInviteEmail(to, name, inviteToken string) error { return nil }
func (m *mockMailer) SendPaymentSuccessEmail(to, name, planName string, amountCents int, currency string) error {
	return nil
}
func (m *mockMailer) SendPaymentFailedEmail(to, name, planName, reason string) error { return nil }
func (m *mockMailer) SendUsageWarningEmail(to, name, resource string, current, limit, pct int) error {
	return nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestListPlans(t *testing.T) {
	plans := []entity.Plan{
		{Name: "free", DisplayName: "Free", MaxContacts: 100, MaxMembers: 3, MaxChatResponders: 1},
		{Name: "pro", DisplayName: "Pro", MaxContacts: 1000, MaxMembers: 10, MaxChatResponders: 5},
		{Name: "business", DisplayName: "Business", MaxContacts: 10000, MaxMembers: 50, MaxChatResponders: 25},
	}
	svc := NewPlanService(&mockPlanRepo{plans: plans}, &mockSubRepo{}, &mockUserRepo{}, &mockMailer{}, testLogger())

	result, err := svc.ListPlans(context.Background())
	if err != nil {
		t.Fatalf("ListPlans() error = %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("ListPlans() returned %d plans, want 3", len(result))
	}
	if result[0].Name != "free" {
		t.Errorf("first plan name = %q, want %q", result[0].Name, "free")
	}
}

func TestGetOrganizationUsage(t *testing.T) {
	orgID := uuid.New()
	now := time.Now()
	freePlan := &entity.Plan{
		Name: "free", DisplayName: "Free",
		MaxContacts: 100, MaxMembers: 3, MaxChatResponders: 1,
	}
	freePlan.ID = 1

	sub := &entity.Subscription{
		OrganizationID:     orgID,
		PlanID:             1,
		Status:             entity.SubscriptionActive,
		CurrentPeriodStart: &now,
	}

	planRepo := &mockPlanRepo{
		planByID:       freePlan,
		contactCount:   42,
		memberCount:    2,
		responderCount: 2,
	}
	subRepo := &mockSubRepo{sub: sub}

	svc := NewPlanService(planRepo, subRepo, &mockUserRepo{}, &mockMailer{}, testLogger())

	usage, err := svc.GetOrganizationUsage(context.Background(), orgID)
	if err != nil {
		t.Fatalf("GetOrganizationUsage() error = %v", err)
	}
	if usage.Plan.Name != "free" {
		t.Errorf("usage plan name = %q, want %q", usage.Plan.Name, "free")
	}
	if usage.Usage.Contacts.Current != 42 {
		t.Errorf("contacts current = %d, want 42", usage.Usage.Contacts.Current)
	}
	if usage.Usage.Contacts.Limit != 100 {
		t.Errorf("contacts limit = %d, want 100", usage.Usage.Contacts.Limit)
	}
	if usage.Subscription.Status != "active" {
		t.Errorf("subscription status = %q, want %q", usage.Subscription.Status, "active")
	}
}

func TestTransitionSubscription_ValidTransition(t *testing.T) {
	orgID := uuid.New()
	sub := &entity.Subscription{
		OrganizationID: orgID,
		Status:         entity.SubscriptionActive,
	}
	sub.ID = 1

	subRepo := &mockSubRepo{sub: sub}
	svc := NewPlanService(&mockPlanRepo{}, subRepo, &mockUserRepo{}, &mockMailer{}, testLogger())

	// active -> past_due is valid
	err := svc.TransitionSubscription(context.Background(), orgID, entity.SubscriptionPastDue)
	if err != nil {
		t.Fatalf("TransitionSubscription(active -> past_due) error = %v", err)
	}
}

func TestTransitionSubscription_InvalidTransition(t *testing.T) {
	orgID := uuid.New()
	sub := &entity.Subscription{
		OrganizationID: orgID,
		Status:         entity.SubscriptionActive,
	}
	sub.ID = 1

	subRepo := &mockSubRepo{sub: sub}
	svc := NewPlanService(&mockPlanRepo{}, subRepo, &mockUserRepo{}, &mockMailer{}, testLogger())

	// active -> expired is NOT valid
	err := svc.TransitionSubscription(context.Background(), orgID, entity.SubscriptionExpired)
	if err == nil {
		t.Fatal("TransitionSubscription(active -> expired) expected error, got nil")
	}
	if err != ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}
