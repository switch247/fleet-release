package unit_tests

import (
	"testing"
	"time"

	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/store"
)

func newStore() *store.MemoryStore {
	return store.NewMemoryStore()
}

// ---------------------------------------------------------------------------
// User CRUD
// ---------------------------------------------------------------------------

func TestSaveAndGetUserByID(t *testing.T) {
	st := newStore()
	u := models.User{ID: "u1", Username: "alice", Email: "alice@test.com", Roles: []models.Role{models.RoleCustomer}}
	st.SaveUser(u)
	got, ok := st.GetUserByID("u1")
	if !ok {
		t.Fatal("expected user to be found by ID")
	}
	if got.Username != "alice" {
		t.Fatalf("expected username alice, got %s", got.Username)
	}
}

func TestGetUserByUsername(t *testing.T) {
	st := newStore()
	u := models.User{ID: "u2", Username: "bob", Email: "bob@test.com", Roles: []models.Role{models.RoleProvider}}
	st.SaveUser(u)
	got, ok := st.GetUserByUsername("bob")
	if !ok {
		t.Fatal("expected user to be found by username")
	}
	if got.ID != "u2" {
		t.Fatalf("expected ID u2, got %s", got.ID)
	}
}

func TestGetUserByID_Missing(t *testing.T) {
	st := newStore()
	_, ok := st.GetUserByID("nonexistent")
	if ok {
		t.Fatal("expected false for missing user")
	}
}

func TestGetUserByUsername_Missing(t *testing.T) {
	st := newStore()
	_, ok := st.GetUserByUsername("nobody")
	if ok {
		t.Fatal("expected false for missing username")
	}
}

func TestListUsers(t *testing.T) {
	st := newStore()
	st.SaveUser(models.User{ID: "u1", Username: "alice", Roles: []models.Role{models.RoleCustomer}})
	st.SaveUser(models.User{ID: "u2", Username: "bob", Roles: []models.Role{models.RoleProvider}})
	users := st.ListUsers()
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestDeleteUser(t *testing.T) {
	st := newStore()
	st.SaveUser(models.User{ID: "u1", Username: "alice", Roles: []models.Role{models.RoleCustomer}})
	st.DeleteUser("u1")
	_, ok := st.GetUserByID("u1")
	if ok {
		t.Fatal("expected user to be deleted by ID")
	}
	_, ok = st.GetUserByUsername("alice")
	if ok {
		t.Fatal("expected username index to be cleaned up after delete")
	}
}

func TestDeleteUser_Idempotent(t *testing.T) {
	st := newStore()
	// Deleting a non-existent user should not panic
	st.DeleteUser("nonexistent")
}

func TestUsernameExists(t *testing.T) {
	st := newStore()
	if st.UsernameExists("carol") {
		t.Fatal("expected false before save")
	}
	st.SaveUser(models.User{ID: "u3", Username: "carol"})
	if !st.UsernameExists("carol") {
		t.Fatal("expected true after save")
	}
}

func TestSaveUserOverwrite(t *testing.T) {
	st := newStore()
	st.SaveUser(models.User{ID: "u1", Username: "alice", Email: "alice@v1.com"})
	st.SaveUser(models.User{ID: "u1", Username: "alice", Email: "alice@v2.com"})
	got, _ := st.GetUserByID("u1")
	if got.Email != "alice@v2.com" {
		t.Fatalf("expected updated email, got %s", got.Email)
	}
}

// ---------------------------------------------------------------------------
// HasAdminExcluding
// ---------------------------------------------------------------------------

func TestHasAdminExcluding_NoAdmin(t *testing.T) {
	st := newStore()
	st.SaveUser(models.User{ID: "u1", Username: "alice", Roles: []models.Role{models.RoleCustomer}})
	if st.HasAdminExcluding("") {
		t.Fatal("expected no admin")
	}
}

func TestHasAdminExcluding_HasAdmin(t *testing.T) {
	st := newStore()
	st.SaveUser(models.User{ID: "u1", Username: "admin1", Roles: []models.Role{models.RoleAdmin}})
	st.SaveUser(models.User{ID: "u2", Username: "admin2", Roles: []models.Role{models.RoleAdmin}})
	if !st.HasAdminExcluding("u2") {
		t.Fatal("expected admin u1 to remain")
	}
}

func TestHasAdminExcluding_OnlyAdmin(t *testing.T) {
	st := newStore()
	st.SaveUser(models.User{ID: "u1", Username: "sole-admin", Roles: []models.Role{models.RoleAdmin}})
	if st.HasAdminExcluding("u1") {
		t.Fatal("expected false when sole admin is excluded")
	}
}

// ---------------------------------------------------------------------------
// Sessions & AuthEvents
// ---------------------------------------------------------------------------

func TestSaveAndGetSession(t *testing.T) {
	st := newStore()
	sess := models.Session{
		ID:          "sess1",
		UserID:      "u1",
		IssuedAt:    time.Now().UTC(),
		LastSeenAt:  time.Now().UTC(),
		AbsoluteExp: time.Now().UTC().Add(12 * time.Hour),
	}
	st.SaveSession(sess)
	got, ok := st.GetSession("sess1")
	if !ok {
		t.Fatal("expected session to be found")
	}
	if got.UserID != "u1" {
		t.Fatalf("expected userID u1, got %s", got.UserID)
	}
}

func TestGetSession_Missing(t *testing.T) {
	st := newStore()
	_, ok := st.GetSession("nosuch")
	if ok {
		t.Fatal("expected false for missing session")
	}
}

func TestSaveAndListAuthEvents(t *testing.T) {
	st := newStore()
	st.SaveAuthEvent(models.AuthEvent{ID: "e1", UserID: "u1", EventType: "login_ok"})
	st.SaveAuthEvent(models.AuthEvent{ID: "e2", UserID: "u1", EventType: "login_fail"})
	events := st.ListAuthEventsByUser("u1", 10)
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
}

func TestListAuthEvents_LimitApplied(t *testing.T) {
	st := newStore()
	for i := 0; i < 5; i++ {
		st.SaveAuthEvent(models.AuthEvent{ID: "e", UserID: "u1", EventType: "login_ok"})
	}
	events := st.ListAuthEventsByUser("u1", 3)
	if len(events) != 3 {
		t.Fatalf("expected 3 events with limit, got %d", len(events))
	}
}

func TestListAuthEvents_AnonymousUser(t *testing.T) {
	st := newStore()
	st.SaveAuthEvent(models.AuthEvent{ID: "anon1", UserID: "", EventType: "login_fail"})
	events := st.ListAuthEventsByUser("anonymous", 10)
	if len(events) != 1 {
		t.Fatalf("expected 1 anonymous event, got %d", len(events))
	}
}
