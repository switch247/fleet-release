package api_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleetlease/backend/pkg/public"
)

func loginForEndpoint(t *testing.T, e http.Handler, username, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed %d %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	return resp.Token
}

func TestFrontendCriticalEndpointsExist(t *testing.T) {
	e := public.BuildSeededRouterForTests()
	token := loginForEndpoint(t, e, "customer", "Customer1234!")

	for _, path := range []string{"/api/v1/stats/summary", "/api/v1/bookings"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s got %d body=%s", path, rec.Code, rec.Body.String())
		}
	}

	estimateBody, _ := json.Marshal(map[string]interface{}{
		"listingId": "11111111-1111-1111-1111-111111111111",
		"startAt":   "2026-03-28T09:00:00Z",
		"endAt":     "2026-03-28T11:00:00Z",
		"odoStart":  10,
		"odoEnd":    25,
	})
	estimateReq := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/estimate", bytes.NewReader(estimateBody))
	estimateReq.Header.Set("Content-Type", "application/json")
	estimateReq.Header.Set("Authorization", "Bearer "+token)
	estimateRec := httptest.NewRecorder()
	e.ServeHTTP(estimateRec, estimateReq)
	if estimateRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /api/v1/bookings/estimate got %d body=%s", estimateRec.Code, estimateRec.Body.String())
	}
}

func TestAdminCategoryAndListingCRUD(t *testing.T) {
	e := public.BuildSeededRouterForTests()
	adminToken := loginForEndpoint(t, e, "admin", "Admin1234!Pass")

	usersReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	usersReq.Header.Set("Authorization", "Bearer "+adminToken)
	usersRec := httptest.NewRecorder()
	e.ServeHTTP(usersRec, usersReq)
	if usersRec.Code != http.StatusOK {
		t.Fatalf("users list failed: %d", usersRec.Code)
	}
	var users []struct {
		ID       string   `json:"id"`
		Username string   `json:"username"`
		Roles    []string `json:"roles"`
	}
	_ = json.Unmarshal(usersRec.Body.Bytes(), &users)
	providerID := ""
	for _, user := range users {
		if user.Username == "provider" {
			providerID = user.ID
		}
	}
	if providerID == "" {
		t.Fatalf("provider user not found")
	}

	createCategoryBody, _ := json.Marshal(map[string]string{"name": "SUV"})
	createCategoryReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/categories", bytes.NewReader(createCategoryBody))
	createCategoryReq.Header.Set("Content-Type", "application/json")
	createCategoryReq.Header.Set("Authorization", "Bearer "+adminToken)
	createCategoryRec := httptest.NewRecorder()
	e.ServeHTTP(createCategoryRec, createCategoryReq)
	if createCategoryRec.Code != http.StatusCreated {
		t.Fatalf("create category failed: %d %s", createCategoryRec.Code, createCategoryRec.Body.String())
	}
	var category struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(createCategoryRec.Body.Bytes(), &category)

	patchCategoryBody, _ := json.Marshal(map[string]string{"name": "SUV Updated"})
	patchCategoryReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/categories/"+category.ID, bytes.NewReader(patchCategoryBody))
	patchCategoryReq.Header.Set("Content-Type", "application/json")
	patchCategoryReq.Header.Set("Authorization", "Bearer "+adminToken)
	patchCategoryRec := httptest.NewRecorder()
	e.ServeHTTP(patchCategoryRec, patchCategoryReq)
	if patchCategoryRec.Code != http.StatusOK {
		t.Fatalf("patch category failed: %d %s", patchCategoryRec.Code, patchCategoryRec.Body.String())
	}

	createListingBody, _ := json.Marshal(map[string]interface{}{
		"categoryId":    category.ID,
		"providerId":    providerID,
		"spu":           "SPU-CRUD",
		"sku":           "SKU-CRUD",
		"name":          "CRUD Listing",
		"includedMiles": 10,
		"deposit":       55,
		"available":     true,
	})
	createListingReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/listings", bytes.NewReader(createListingBody))
	createListingReq.Header.Set("Content-Type", "application/json")
	createListingReq.Header.Set("Authorization", "Bearer "+adminToken)
	createListingRec := httptest.NewRecorder()
	e.ServeHTTP(createListingRec, createListingReq)
	if createListingRec.Code != http.StatusCreated {
		t.Fatalf("create listing failed: %d %s", createListingRec.Code, createListingRec.Body.String())
	}
	var listing struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(createListingRec.Body.Bytes(), &listing)

	patchListingBody, _ := json.Marshal(map[string]interface{}{"name": "CRUD Listing Updated", "deposit": 65, "available": false})
	patchListingReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/listings/"+listing.ID, bytes.NewReader(patchListingBody))
	patchListingReq.Header.Set("Content-Type", "application/json")
	patchListingReq.Header.Set("Authorization", "Bearer "+adminToken)
	patchListingRec := httptest.NewRecorder()
	e.ServeHTTP(patchListingRec, patchListingReq)
	if patchListingRec.Code != http.StatusOK {
		t.Fatalf("patch listing failed: %d %s", patchListingRec.Code, patchListingRec.Body.String())
	}

	deleteListingReq := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/listings/"+listing.ID, nil)
	deleteListingReq.Header.Set("Authorization", "Bearer "+adminToken)
	deleteListingRec := httptest.NewRecorder()
	e.ServeHTTP(deleteListingRec, deleteListingReq)
	if deleteListingRec.Code != http.StatusNoContent {
		t.Fatalf("delete listing failed: %d %s", deleteListingRec.Code, deleteListingRec.Body.String())
	}

	deleteCategoryReq := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/categories/"+category.ID, nil)
	deleteCategoryReq.Header.Set("Authorization", "Bearer "+adminToken)
	deleteCategoryRec := httptest.NewRecorder()
	e.ServeHTTP(deleteCategoryRec, deleteCategoryReq)
	if deleteCategoryRec.Code != http.StatusNoContent {
		t.Fatalf("delete category failed: %d %s", deleteCategoryRec.Code, deleteCategoryRec.Body.String())
	}
}

func TestAdminCanUpdateUserRoles(t *testing.T) {
	e := public.BuildSeededRouterForTests()
	adminToken := loginForEndpoint(t, e, "admin", "Admin1234!Pass")

	createUserBody, _ := json.Marshal(map[string]interface{}{
		"username": "ops-user",
		"email":    "ops-user@fleetlease.local",
		"password": "OpsUser1234!A",
		"roles":    []string{"customer"},
	})
	createUserReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(createUserBody))
	createUserReq.Header.Set("Content-Type", "application/json")
	createUserReq.Header.Set("Authorization", "Bearer "+adminToken)
	createUserRec := httptest.NewRecorder()
	e.ServeHTTP(createUserRec, createUserReq)
	if createUserRec.Code != http.StatusCreated {
		t.Fatalf("create user failed: %d %s", createUserRec.Code, createUserRec.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(createUserRec.Body.Bytes(), &created)
	if created.ID == "" {
		t.Fatalf("expected created user ID")
	}

	updateBody, _ := json.Marshal(map[string]interface{}{
		"roles": []string{"provider"},
		"email": "ops-user-updated@fleetlease.local",
	})
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+created.ID, bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+adminToken)
	updateRec := httptest.NewRecorder()
	e.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update user failed: %d %s", updateRec.Code, updateRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	listReq.Header.Set("Authorization", "Bearer "+adminToken)
	listRec := httptest.NewRecorder()
	e.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list users failed: %d %s", listRec.Code, listRec.Body.String())
	}
	var users []struct {
		ID    string   `json:"id"`
		Email string   `json:"email"`
		Roles []string `json:"roles"`
	}
	_ = json.Unmarshal(listRec.Body.Bytes(), &users)
	seen := false
	for _, user := range users {
		if user.ID == created.ID {
			seen = true
			if user.Email != "ops-user-updated@fleetlease.local" {
				t.Fatalf("expected updated email, got %s", user.Email)
			}
			if len(user.Roles) != 1 || user.Roles[0] != "provider" {
				t.Fatalf("expected provider role, got %+v", user.Roles)
			}
		}
	}
	if !seen {
		t.Fatalf("updated user not found in list")
	}
}

func TestCategoryTreeView(t *testing.T) {
	e := public.BuildSeededRouterForTests()
	adminToken := loginForEndpoint(t, e, "admin", "Admin1234!Pass")
	customerToken := loginForEndpoint(t, e, "customer", "Customer1234!")

	parentBody, _ := json.Marshal(map[string]string{"name": "Vehicles"})
	parentReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/categories", bytes.NewReader(parentBody))
	parentReq.Header.Set("Content-Type", "application/json")
	parentReq.Header.Set("Authorization", "Bearer "+adminToken)
	parentRec := httptest.NewRecorder()
	e.ServeHTTP(parentRec, parentReq)
	if parentRec.Code != http.StatusCreated {
		t.Fatalf("create parent category failed: %d %s", parentRec.Code, parentRec.Body.String())
	}
	var parent struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(parentRec.Body.Bytes(), &parent)

	childBody, _ := json.Marshal(map[string]string{"name": "SUV", "parentId": parent.ID})
	childReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/categories", bytes.NewReader(childBody))
	childReq.Header.Set("Content-Type", "application/json")
	childReq.Header.Set("Authorization", "Bearer "+adminToken)
	childRec := httptest.NewRecorder()
	e.ServeHTTP(childRec, childReq)
	if childRec.Code != http.StatusCreated {
		t.Fatalf("create child category failed: %d %s", childRec.Code, childRec.Body.String())
	}

	treeReq := httptest.NewRequest(http.MethodGet, "/api/v1/categories?view=tree", nil)
	treeReq.Header.Set("Authorization", "Bearer "+customerToken)
	treeRec := httptest.NewRecorder()
	e.ServeHTTP(treeRec, treeReq)
	if treeRec.Code != http.StatusOK {
		t.Fatalf("tree categories failed: %d %s", treeRec.Code, treeRec.Body.String())
	}
	var nodes []struct {
		ID       string `json:"id"`
		Children []struct {
			ID string `json:"id"`
		} `json:"children"`
	}
	_ = json.Unmarshal(treeRec.Body.Bytes(), &nodes)
	foundParentWithChild := false
	for _, node := range nodes {
		if node.ID != parent.ID {
			continue
		}
		if len(node.Children) == 1 {
			foundParentWithChild = true
		}
	}
	if !foundParentWithChild {
		t.Fatalf("expected parent category to include child in tree response")
	}
}
