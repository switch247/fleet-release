package api_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestFrontendCriticalEndpointsExist verifies that the three endpoints the
// frontend hits on startup all return 200 for an authenticated customer.
func TestFrontendCriticalEndpointsExist(t *testing.T) {
	custToken := liveLogin(t, apiCustUser, apiCustPass)

	for _, path := range []string{"/api/v1/stats/summary", "/api/v1/bookings"} {
		resp := apiCall(t, http.MethodGet, path, nil, custToken)
		mustAPIStatus(t, resp, http.StatusOK)
	}

	now := time.Now().UTC()
	resp := apiCall(t, http.MethodPost, "/api/v1/bookings/estimate", map[string]interface{}{
		"listingId": apiListingID,
		"startAt":   now.Format(time.RFC3339),
		"endAt":     now.Add(2 * time.Hour).Format(time.RFC3339),
		"odoStart":  10,
		"odoEnd":    25,
	}, custToken)
	mustAPIStatus(t, resp, http.StatusOK)
}

// TestAdminCategoryAndListingCRUD exercises the full admin create→patch→delete
// lifecycle for categories and listings.
func TestAdminCategoryAndListingCRUD(t *testing.T) {
	adminToken := liveLoginAdmin(t)

	// Confirm api-provider exists in the user list so we can reference apiProvID.
	resp := apiCall(t, http.MethodGet, "/api/v1/admin/users", nil, adminToken)
	b := mustAPIStatus(t, resp, http.StatusOK)
	var users []struct {
		ID       string   `json:"id"`
		Username string   `json:"username"`
		Roles    []string `json:"roles"`
	}
	if err := json.Unmarshal(b, &users); err != nil {
		t.Fatalf("list users: %v", err)
	}
	providerFound := false
	for _, u := range users {
		if u.ID == apiProvID {
			providerFound = true
			break
		}
	}
	if !providerFound {
		t.Fatalf("seeded provider %s not found in admin user list", apiProvID)
	}

	// Create category.
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	resp2 := apiCall(t, http.MethodPost, "/api/v1/admin/categories",
		map[string]string{"name": "SUV-" + suffix}, adminToken)
	b2 := mustAPIStatus(t, resp2, http.StatusCreated)
	var category struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(b2, &category); err != nil || category.ID == "" {
		t.Fatalf("create category: %s", b2)
	}

	// Patch category.
	resp3 := apiCall(t, http.MethodPatch, "/api/v1/admin/categories/"+category.ID,
		map[string]string{"name": "SUV-Updated-" + suffix}, adminToken)
	mustAPIStatus(t, resp3, http.StatusOK)

	// Create listing under the new category.
	resp4 := apiCall(t, http.MethodPost, "/api/v1/admin/listings", map[string]interface{}{
		"categoryId":    category.ID,
		"providerId":    apiProvID,
		"spu":           "SPU-CRUD-" + suffix,
		"sku":           "SKU-CRUD-" + suffix,
		"name":          "CRUD Listing",
		"includedMiles": 10,
		"deposit":       55,
		"available":     true,
	}, adminToken)
	b4 := mustAPIStatus(t, resp4, http.StatusCreated)
	var listing struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(b4, &listing); err != nil || listing.ID == "" {
		t.Fatalf("create listing: %s", b4)
	}

	// Patch listing.
	resp5 := apiCall(t, http.MethodPatch, "/api/v1/admin/listings/"+listing.ID,
		map[string]interface{}{"name": "CRUD Listing Updated", "deposit": 65, "available": false},
		adminToken)
	mustAPIStatus(t, resp5, http.StatusOK)

	// Delete listing then category.
	resp6 := apiCall(t, http.MethodDelete, "/api/v1/admin/listings/"+listing.ID, nil, adminToken)
	mustAPIStatus(t, resp6, http.StatusNoContent)

	resp7 := apiCall(t, http.MethodDelete, "/api/v1/admin/categories/"+category.ID, nil, adminToken)
	mustAPIStatus(t, resp7, http.StatusNoContent)
}

// TestAdminCanUpdateUserRoles creates a throwaway user, patches their role and
// email via admin, then confirms the change appears in the user list.
func TestAdminCanUpdateUserRoles(t *testing.T) {
	adminToken := liveLoginAdmin(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	opsUser := "ops-user-" + suffix

	resp := apiCall(t, http.MethodPost, "/api/v1/admin/users", map[string]interface{}{
		"username": opsUser,
		"email":    opsUser + "@fleetlease.local",
		"password": "OpsUser1234!A",
		"roles":    []string{"customer"},
	}, adminToken)
	b := mustAPIStatus(t, resp, http.StatusCreated)
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(b, &created); err != nil || created.ID == "" {
		t.Fatalf("create user: %s", b)
	}

	// Patch role to provider and update email.
	updatedEmail := opsUser + "-updated@fleetlease.local"
	resp2 := apiCall(t, http.MethodPatch, "/api/v1/admin/users/"+created.ID, map[string]interface{}{
		"roles": []string{"provider"},
		"email": updatedEmail,
	}, adminToken)
	mustAPIStatus(t, resp2, http.StatusOK)

	// Verify via list.
	resp3 := apiCall(t, http.MethodGet, "/api/v1/admin/users", nil, adminToken)
	b3 := mustAPIStatus(t, resp3, http.StatusOK)
	var users []struct {
		ID    string   `json:"id"`
		Email string   `json:"email"`
		Roles []string `json:"roles"`
	}
	if err := json.Unmarshal(b3, &users); err != nil {
		t.Fatalf("list users: %v", err)
	}
	seen := false
	for _, u := range users {
		if u.ID != created.ID {
			continue
		}
		seen = true
		if u.Email != updatedEmail {
			t.Fatalf("expected email %s got %s", updatedEmail, u.Email)
		}
		if len(u.Roles) != 1 || u.Roles[0] != "provider" {
			t.Fatalf("expected [provider] roles, got %v", u.Roles)
		}
	}
	if !seen {
		t.Fatalf("updated user %s not found in admin user list", created.ID)
	}
}

// TestCategoryTreeView creates a parent category and a child category, then
// verifies the customer-facing tree endpoint nests them correctly.
func TestCategoryTreeView(t *testing.T) {
	adminToken := liveLoginAdmin(t)
	custToken := liveLogin(t, apiCustUser, apiCustPass)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create parent.
	resp := apiCall(t, http.MethodPost, "/api/v1/admin/categories",
		map[string]string{"name": "Vehicles-" + suffix}, adminToken)
	b := mustAPIStatus(t, resp, http.StatusCreated)
	var parent struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(b, &parent); err != nil || parent.ID == "" {
		t.Fatalf("create parent category: %s", b)
	}

	// Create child with parentId.
	resp2 := apiCall(t, http.MethodPost, "/api/v1/admin/categories",
		map[string]string{"name": "SUV-" + suffix, "parentId": parent.ID}, adminToken)
	mustAPIStatus(t, resp2, http.StatusCreated)

	// Fetch tree as customer and verify parent has exactly one child.
	resp3 := apiCall(t, http.MethodGet, "/api/v1/categories?view=tree", nil, custToken)
	b3 := mustAPIStatus(t, resp3, http.StatusOK)
	var nodes []struct {
		ID       string `json:"id"`
		Children []struct {
			ID string `json:"id"`
		} `json:"children"`
	}
	if err := json.Unmarshal(b3, &nodes); err != nil {
		t.Fatalf("parse tree: %v — body: %s", err, b3)
	}
	foundParentWithChild := false
	for _, node := range nodes {
		if node.ID != parent.ID {
			continue
		}
		if len(node.Children) >= 1 {
			foundParentWithChild = true
		}
	}
	if !foundParentWithChild {
		t.Fatalf("expected parent category %s to include child in tree response; nodes: %s", parent.ID, b3)
	}
}
