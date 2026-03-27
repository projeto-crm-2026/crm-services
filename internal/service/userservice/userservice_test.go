package userservice

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/projeto-crm-2026/crm-services/internal/config"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type mockUserRepo struct {
	users       map[string]*entity.User // keyed by email
	usersById   map[uint]*entity.User
	insertedOrg *uuid.UUID
	err         error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:     make(map[string]*entity.User),
		usersById: make(map[uint]*entity.User),
	}
}

func (m *mockUserRepo) Insert(ctx context.Context, name, email, passwordHash string, organizationID uuid.UUID) (*entity.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.insertedOrg = &organizationID
	user := &entity.User{
		Name:           name,
		Email:          email,
		PasswordHash:   passwordHash,
		OrganizationID: &organizationID,
		Role:           entity.RoleAdmin,
		Status:         entity.StatusActive,
	}
	user.ID = uint(len(m.usersById) + 1)
	user.UUID = uuid.New()
	m.users[email] = user
	m.usersById[user.ID] = user
	return user, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	if u, ok := m.usersById[id]; ok {
		return u, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockUserRepo) GetByInviteToken(ctx context.Context, token string) (*entity.User, error) {
	for _, u := range m.users {
		if u.InviteToken != nil && *u.InviteToken == token {
			return u, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *mockUserRepo) InsertPending(ctx context.Context, name, email, inviteToken string, inviteExpiry time.Time, organizationID uuid.UUID, invitedBy uint) (*entity.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	user := &entity.User{
		Name:           name,
		Email:          email,
		OrganizationID: &organizationID,
		Role:           entity.RoleMember,
		Status:         entity.StatusPending,
		InviteToken:    &inviteToken,
		InviteExpiry:   &inviteExpiry,
		InvitedBy:      &invitedBy,
	}
	user.ID = uint(len(m.usersById) + 1)
	user.UUID = uuid.New()
	m.users[email] = user
	m.usersById[user.ID] = user
	return user, nil
}

func (m *mockUserRepo) ActivateUser(ctx context.Context, userID uint, passwordHash string) error {
	if u, ok := m.usersById[userID]; ok {
		u.Status = entity.StatusActive
		u.PasswordHash = passwordHash
		u.InviteToken = nil
		u.InviteExpiry = nil
		return nil
	}
	return fmt.Errorf("user not found")
}

func (m *mockUserRepo) UpdateRoleID(ctx context.Context, userID, roleID uint) error { return nil }

func (m *mockUserRepo) DeactivateUser(ctx context.Context, userID uint) error {
	if u, ok := m.usersById[userID]; ok {
		u.Status = entity.StatusDeactivated
		return nil
	}
	return fmt.Errorf("user not found")
}

func (m *mockUserRepo) ReactivateUser(ctx context.Context, userID uint) error {
	if u, ok := m.usersById[userID]; ok {
		u.Status = entity.StatusActive
		return nil
	}
	return fmt.Errorf("user not found")
}

func (m *mockUserRepo) RemoveMember(ctx context.Context, userID uint) error {
	if u, ok := m.usersById[userID]; ok {
		u.OrganizationID = nil
		u.DeletedAt.Valid = true
		return nil
	}
	return fmt.Errorf("user not found")
}

func (m *mockUserRepo) ListByOrganization(ctx context.Context, organizationID uuid.UUID) ([]entity.User, error) {
	var result []entity.User
	for _, u := range m.usersById {
		if u.OrganizationID != nil && *u.OrganizationID == organizationID {
			result = append(result, *u)
		}
	}
	return result, nil
}

func (m *mockUserRepo) ListByOrganizationWithRole(ctx context.Context, organizationID uuid.UUID) ([]entity.User, error) {
	return m.ListByOrganization(ctx, organizationID)
}

func (m *mockUserRepo) GetOrgOwnerEmail(ctx context.Context, orgID uuid.UUID) (string, string, error) {
	return "Owner", "owner@test.com", nil
}

type mockOrgRepo struct {
	createdOrg *entity.Organization
	err        error
}

func (m *mockOrgRepo) Create(ctx context.Context, org *entity.Organization) (*entity.Organization, error) {
	if m.err != nil {
		return nil, m.err
	}
	org.UUID = uuid.New()
	org.ID = 1
	m.createdOrg = org
	return org, nil
}

func (m *mockOrgRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error) {
	return nil, nil
}
func (m *mockOrgRepo) GetBySlug(ctx context.Context, slug string) (*entity.Organization, error) {
	return nil, nil
}
func (m *mockOrgRepo) Update(ctx context.Context, org *entity.Organization) error { return nil }
func (m *mockOrgRepo) Delete(ctx context.Context, id uuid.UUID) error             { return nil }
func (m *mockOrgRepo) SoftDelete(ctx context.Context, id uuid.UUID) error         { return nil }
func (m *mockOrgRepo) Restore(ctx context.Context, id uuid.UUID) error            { return nil }

type mockPlanRepo struct {
	freePlan *entity.Plan
	err      error
}

func (m *mockPlanRepo) GetAll(ctx context.Context) ([]entity.Plan, error) { return nil, m.err }
func (m *mockPlanRepo) GetByID(ctx context.Context, id uint) (*entity.Plan, error) {
	return m.freePlan, m.err
}
func (m *mockPlanRepo) GetByName(ctx context.Context, name string) (*entity.Plan, error) {
	if m.freePlan != nil && m.freePlan.Name == name {
		return m.freePlan, nil
	}
	return nil, fmt.Errorf("plan not found")
}
func (m *mockPlanRepo) CountContacts(ctx context.Context, orgID uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockPlanRepo) CountMembers(ctx context.Context, orgID uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockPlanRepo) CountChatResponders(ctx context.Context, orgID uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockPlanRepo) GetByMpPlanID(ctx context.Context, mpPlanID string) (*entity.Plan, error) {
	return nil, nil
}
func (m *mockPlanRepo) UpdateMpPlanID(ctx context.Context, planID uint, mpPlanID string) error {
	return nil
}

type mockSubRepo struct {
	created *entity.Subscription
	err     error
}

func (m *mockSubRepo) Create(ctx context.Context, sub *entity.Subscription) (*entity.Subscription, error) {
	if m.err != nil {
		return nil, m.err
	}
	sub.ID = 1
	sub.UUID = uuid.New()
	m.created = sub
	return sub, nil
}
func (m *mockSubRepo) GetByOrganizationID(ctx context.Context, orgID uuid.UUID) (*entity.Subscription, error) {
	if m.created != nil {
		return m.created, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockSubRepo) GetByID(ctx context.Context, id uint) (*entity.Subscription, error) {
	return nil, nil
}
func (m *mockSubRepo) UpdateStatus(ctx context.Context, id uint, status entity.SubscriptionStatus) error {
	return nil
}
func (m *mockSubRepo) Update(ctx context.Context, sub *entity.Subscription) error { return nil }
func (m *mockSubRepo) GetByMpSubscriptionID(ctx context.Context, mpSubID string) (*entity.Subscription, error) {
	return nil, nil
}

type mockMailer struct {
	invitesSent int
}

func (m *mockMailer) SendInviteEmail(to, name, inviteToken string) error {
	m.invitesSent++
	return nil
}
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

func newTestService(userRepo *mockUserRepo, orgRepo *mockOrgRepo, planRepo *mockPlanRepo, subRepo *mockSubRepo, mail *mockMailer) UserService {
	return NewUserService(userRepo, orgRepo, planRepo, subRepo, &config.JWTConfig{JWTSecret: "test-secret-key-32-chars-long!!!"}, mail, testLogger())
}

func TestRegisterAssignsFreePlan(t *testing.T) {
	freePlan := &entity.Plan{Name: "free", DisplayName: "Free", MaxContacts: 100, MaxMembers: 3, MaxChatResponders: 1}
	freePlan.ID = 1

	userRepo := newMockUserRepo()
	orgRepo := &mockOrgRepo{}
	planRepo := &mockPlanRepo{freePlan: freePlan}
	subRepo := &mockSubRepo{}
	mail := &mockMailer{}

	svc := newTestService(userRepo, orgRepo, planRepo, subRepo, mail)

	token, user, err := svc.RegisterUser(context.Background(), "John", "john@test.com", "password123", "My Org")
	if err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}

	if orgRepo.createdOrg == nil {
		t.Fatal("expected organization to be created")
	}
	if orgRepo.createdOrg.PlanID == nil || *orgRepo.createdOrg.PlanID != freePlan.ID {
		t.Errorf("organization PlanID = %v, want %d", orgRepo.createdOrg.PlanID, freePlan.ID)
	}

	if subRepo.created == nil {
		t.Fatal("expected subscription to be created")
	}
	if subRepo.created.PlanID != freePlan.ID {
		t.Errorf("subscription PlanID = %d, want %d", subRepo.created.PlanID, freePlan.ID)
	}
	if subRepo.created.Status != entity.SubscriptionActive {
		t.Errorf("subscription status = %q, want %q", subRepo.created.Status, entity.SubscriptionActive)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	userRepo := newMockUserRepo()
	userRepo.users["john@test.com"] = &entity.User{Email: "john@test.com"}

	freePlan := &entity.Plan{Name: "free"}
	freePlan.ID = 1

	svc := newTestService(userRepo, &mockOrgRepo{}, &mockPlanRepo{freePlan: freePlan}, &mockSubRepo{}, &mockMailer{})

	_, _, err := svc.RegisterUser(context.Background(), "John", "john@test.com", "password123", "My Org")
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
}

func TestLoginDeactivatedUser(t *testing.T) {
	userRepo := newMockUserRepo()
	orgID := uuid.New()
	userRepo.users["deactivated@test.com"] = &entity.User{
		Email:          "deactivated@test.com",
		OrganizationID: &orgID,
		Status:         entity.StatusDeactivated,
		PasswordHash:   "irrelevant",
	}

	svc := newTestService(userRepo, &mockOrgRepo{}, &mockPlanRepo{}, &mockSubRepo{}, &mockMailer{})

	_, _, err := svc.LoginUser(context.Background(), "deactivated@test.com", "password123")
	if err == nil {
		t.Fatal("expected error for deactivated user, got nil")
	}
	if err.Error() != "account has been deactivated, contact your organization admin" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoginPendingUser(t *testing.T) {
	userRepo := newMockUserRepo()
	orgID := uuid.New()
	userRepo.users["pending@test.com"] = &entity.User{
		Email:          "pending@test.com",
		OrganizationID: &orgID,
		Status:         entity.StatusPending,
	}

	svc := newTestService(userRepo, &mockOrgRepo{}, &mockPlanRepo{}, &mockSubRepo{}, &mockMailer{})

	_, _, err := svc.LoginUser(context.Background(), "pending@test.com", "password123")
	if err == nil {
		t.Fatal("expected error for pending user, got nil")
	}
	if err.Error() != "account is pending activation, check your email" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestInviteUser(t *testing.T) {
	userRepo := newMockUserRepo()
	orgID := uuid.New()
	admin := &entity.User{
		Name:           "Admin",
		Email:          "admin@test.com",
		OrganizationID: &orgID,
		Role:           entity.RoleAdmin,
		Status:         entity.StatusActive,
	}
	admin.ID = 1
	userRepo.usersById[1] = admin
	userRepo.users["admin@test.com"] = admin

	mail := &mockMailer{}
	svc := newTestService(userRepo, &mockOrgRepo{}, &mockPlanRepo{}, &mockSubRepo{}, mail)

	invited, err := svc.InviteUser(context.Background(), 1, "New Member", "newmember@test.com")
	if err != nil {
		t.Fatalf("InviteUser() error = %v", err)
	}
	if invited.Status != entity.StatusPending {
		t.Errorf("invited user status = %q, want %q", invited.Status, entity.StatusPending)
	}
	if invited.InviteToken == nil {
		t.Error("expected invite token to be set")
	}
	if mail.invitesSent != 1 {
		t.Errorf("expected 1 invite email sent, got %d", mail.invitesSent)
	}
}

func TestInviteDuplicateEmail(t *testing.T) {
	userRepo := newMockUserRepo()
	orgID := uuid.New()
	admin := &entity.User{OrganizationID: &orgID, Role: entity.RoleAdmin, Status: entity.StatusActive}
	admin.ID = 1
	userRepo.usersById[1] = admin
	userRepo.users["existing@test.com"] = &entity.User{Email: "existing@test.com"}

	svc := newTestService(userRepo, &mockOrgRepo{}, &mockPlanRepo{}, &mockSubRepo{}, &mockMailer{})

	_, err := svc.InviteUser(context.Background(), 1, "Existing", "existing@test.com")
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
}

func TestAcceptInvite(t *testing.T) {
	userRepo := newMockUserRepo()
	orgID := uuid.New()
	token := "inv_test-token"
	expiry := time.Now().Add(72 * time.Hour)
	pending := &entity.User{
		Name:           "New Member",
		Email:          "new@test.com",
		OrganizationID: &orgID,
		Status:         entity.StatusPending,
		InviteToken:    &token,
		InviteExpiry:   &expiry,
	}
	pending.ID = 1
	userRepo.usersById[1] = pending
	userRepo.users["new@test.com"] = pending

	svc := newTestService(userRepo, &mockOrgRepo{}, &mockPlanRepo{}, &mockSubRepo{}, &mockMailer{})

	user, err := svc.AcceptInvite(context.Background(), token, "newpassword123")
	if err != nil {
		t.Fatalf("AcceptInvite() error = %v", err)
	}
	if user.Status != entity.StatusActive {
		t.Errorf("user status = %q, want %q", user.Status, entity.StatusActive)
	}
}

func TestAcceptInviteExpired(t *testing.T) {
	userRepo := newMockUserRepo()
	orgID := uuid.New()
	token := "inv_expired"
	expiry := time.Now().Add(-1 * time.Hour) // expired
	pending := &entity.User{
		Name:           "Expired",
		Email:          "expired@test.com",
		OrganizationID: &orgID,
		Status:         entity.StatusPending,
		InviteToken:    &token,
		InviteExpiry:   &expiry,
	}
	pending.ID = 1
	userRepo.usersById[1] = pending
	userRepo.users["expired@test.com"] = pending

	svc := newTestService(userRepo, &mockOrgRepo{}, &mockPlanRepo{}, &mockSubRepo{}, &mockMailer{})

	_, err := svc.AcceptInvite(context.Background(), token, "newpassword123")
	if err == nil {
		t.Fatal("expected error for expired invite, got nil")
	}
}

func TestDeactivateMember(t *testing.T) {
	userRepo := newMockUserRepo()
	orgID := uuid.New()
	member := &entity.User{
		OrganizationID: &orgID,
		Status:         entity.StatusActive,
		Role:           entity.RoleMember,
	}
	member.ID = 2
	userRepo.usersById[2] = member

	svc := newTestService(userRepo, &mockOrgRepo{}, &mockPlanRepo{}, &mockSubRepo{}, &mockMailer{})

	err := svc.DeactivateMember(context.Background(), orgID, 2)
	if err != nil {
		t.Fatalf("DeactivateMember() error = %v", err)
	}
	if member.Status != entity.StatusDeactivated {
		t.Errorf("member status = %q, want %q", member.Status, entity.StatusDeactivated)
	}
}

func TestDeactivateAlreadyDeactivated(t *testing.T) {
	userRepo := newMockUserRepo()
	orgID := uuid.New()
	member := &entity.User{OrganizationID: &orgID, Status: entity.StatusDeactivated}
	member.ID = 2
	userRepo.usersById[2] = member

	svc := newTestService(userRepo, &mockOrgRepo{}, &mockPlanRepo{}, &mockSubRepo{}, &mockMailer{})

	err := svc.DeactivateMember(context.Background(), orgID, 2)
	if err == nil {
		t.Fatal("expected error for already deactivated, got nil")
	}
}

func TestReactivateMember(t *testing.T) {
	userRepo := newMockUserRepo()
	orgID := uuid.New()
	member := &entity.User{
		OrganizationID: &orgID,
		Status:         entity.StatusDeactivated,
	}
	member.ID = 2
	userRepo.usersById[2] = member

	svc := newTestService(userRepo, &mockOrgRepo{}, &mockPlanRepo{}, &mockSubRepo{}, &mockMailer{})

	err := svc.ReactivateMember(context.Background(), orgID, 2)
	if err != nil {
		t.Fatalf("ReactivateMember() error = %v", err)
	}
	if member.Status != entity.StatusActive {
		t.Errorf("member status = %q, want %q", member.Status, entity.StatusActive)
	}
}

func TestRemoveMember(t *testing.T) {
	userRepo := newMockUserRepo()
	orgID := uuid.New()
	member := &entity.User{
		OrganizationID: &orgID,
		Status:         entity.StatusActive,
		Role:           entity.RoleMember,
	}
	member.ID = 2
	userRepo.usersById[2] = member

	svc := newTestService(userRepo, &mockOrgRepo{}, &mockPlanRepo{}, &mockSubRepo{}, &mockMailer{})

	err := svc.RemoveMember(context.Background(), orgID, 2)
	if err != nil {
		t.Fatalf("RemoveMember() error = %v", err)
	}
	if member.OrganizationID != nil {
		t.Error("expected organization_id to be nil after removal")
	}
}

func TestRemoveMemberWrongOrg(t *testing.T) {
	userRepo := newMockUserRepo()
	orgID := uuid.New()
	otherOrgID := uuid.New()
	member := &entity.User{OrganizationID: &otherOrgID, Status: entity.StatusActive}
	member.ID = 2
	userRepo.usersById[2] = member

	svc := newTestService(userRepo, &mockOrgRepo{}, &mockPlanRepo{}, &mockSubRepo{}, &mockMailer{})

	err := svc.RemoveMember(context.Background(), orgID, 2)
	if err == nil {
		t.Fatal("expected error for wrong org, got nil")
	}
}
