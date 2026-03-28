package handlers

import (
	"net/http"
	"strings"

	"fleetlease/backend/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) AdminCreateCategory(c echo.Context) error {
	var req struct {
		Name     string `json:"name"`
		ParentID string `json:"parentId"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
	}
	parentID := strings.TrimSpace(req.ParentID)
	if parentID != "" {
		if _, ok := h.Store.GetCategory(parentID); !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "parent category not found"})
		}
	}
	category := models.Category{
		ID:       uuid.NewString(),
		Name:     name,
		ParentID: parentID,
	}
	h.Store.SaveCategory(category)
	h.Logger.Info("admin_category_created", "categoryID", category.ID, "name", category.Name)
	return c.JSON(http.StatusCreated, category)
}

func (h *Handler) AdminListCategories(c echo.Context) error {
	return c.JSON(http.StatusOK, h.Store.ListCategories())
}

func (h *Handler) AdminUpdateCategory(c echo.Context) error {
	categoryID := c.Param("categoryID")
	category, ok := h.Store.GetCategory(categoryID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "category not found"})
	}
	var req struct {
		Name     string  `json:"name"`
		ParentID *string `json:"parentId"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
	}
	category.Name = name
	if req.ParentID != nil {
		parentID := strings.TrimSpace(*req.ParentID)
		if parentID == category.ID {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "category cannot parent itself"})
		}
		if parentID != "" {
			if _, ok := h.Store.GetCategory(parentID); !ok {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "parent category not found"})
			}
		}
		category.ParentID = parentID
	}
	h.Store.SaveCategory(category)
	h.Logger.Info("admin_category_updated", "categoryID", category.ID)
	return c.JSON(http.StatusOK, category)
}

func (h *Handler) AdminDeleteCategory(c echo.Context) error {
	categoryID := c.Param("categoryID")
	if _, ok := h.Store.GetCategory(categoryID); !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "category not found"})
	}
	for _, listing := range h.Store.ListListings() {
		if listing.CategoryID == categoryID {
			return c.JSON(http.StatusConflict, map[string]string{"error": "category has listings and cannot be deleted"})
		}
	}
	for _, category := range h.Store.ListCategories() {
		if category.ParentID == categoryID {
			return c.JSON(http.StatusConflict, map[string]string{"error": "category has child categories and cannot be deleted"})
		}
	}
	h.Store.DeleteCategory(categoryID)
	h.Logger.Info("admin_category_deleted", "categoryID", categoryID)
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AdminCreateListing(c echo.Context) error {
	var req struct {
		CategoryID    string  `json:"categoryId"`
		ProviderID    string  `json:"providerId"`
		SPU           string  `json:"spu"`
		SKU           string  `json:"sku"`
		Name          string  `json:"name"`
		IncludedMiles float64 `json:"includedMiles"`
		Deposit       float64 `json:"deposit"`
		Available     bool    `json:"available"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	req.CategoryID = strings.TrimSpace(req.CategoryID)
	req.ProviderID = strings.TrimSpace(req.ProviderID)
	req.SPU = strings.TrimSpace(req.SPU)
	req.SKU = strings.TrimSpace(req.SKU)
	req.Name = strings.TrimSpace(req.Name)
	if req.CategoryID == "" || req.SPU == "" || req.SKU == "" || req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "categoryId, spu, sku, and name are required"})
	}
	if _, ok := h.Store.GetCategory(req.CategoryID); !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "category not found"})
	}
	if req.ProviderID != "" {
		provider, ok := h.Store.GetUserByID(req.ProviderID)
		if !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "provider not found"})
		}
		if !hasRole(provider.Roles, models.RoleProvider) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "providerId must belong to a provider"})
		}
	}
	listing := models.Listing{
		ID:            uuid.NewString(),
		CategoryID:    req.CategoryID,
		ProviderID:    req.ProviderID,
		SPU:           req.SPU,
		SKU:           req.SKU,
		Name:          req.Name,
		IncludedMiles: req.IncludedMiles,
		Deposit:       req.Deposit,
		Available:     req.Available,
	}
	h.Store.SaveListing(listing)
	h.Logger.Info("admin_listing_created", "listingID", listing.ID, "categoryID", listing.CategoryID)
	return c.JSON(http.StatusCreated, listing)
}

func (h *Handler) AdminListListings(c echo.Context) error {
	return c.JSON(http.StatusOK, h.Store.ListListings())
}

func (h *Handler) AdminUpdateListing(c echo.Context) error {
	listingID := c.Param("listingID")
	listing, ok := h.Store.GetListing(listingID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "listing not found"})
	}
	var req struct {
		CategoryID    *string  `json:"categoryId"`
		ProviderID    *string  `json:"providerId"`
		SPU           *string  `json:"spu"`
		SKU           *string  `json:"sku"`
		Name          *string  `json:"name"`
		IncludedMiles *float64 `json:"includedMiles"`
		Deposit       *float64 `json:"deposit"`
		Available     *bool    `json:"available"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	if req.CategoryID != nil {
		categoryID := strings.TrimSpace(*req.CategoryID)
		if categoryID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "categoryId cannot be empty"})
		}
		if _, ok := h.Store.GetCategory(categoryID); !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "category not found"})
		}
		listing.CategoryID = categoryID
	}
	if req.ProviderID != nil {
		providerID := strings.TrimSpace(*req.ProviderID)
		if providerID != "" {
			provider, ok := h.Store.GetUserByID(providerID)
			if !ok {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "provider not found"})
			}
			if !hasRole(provider.Roles, models.RoleProvider) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "providerId must belong to a provider"})
			}
		}
		listing.ProviderID = providerID
	}
	if req.SPU != nil {
		spu := strings.TrimSpace(*req.SPU)
		if spu == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "spu cannot be empty"})
		}
		listing.SPU = spu
	}
	if req.SKU != nil {
		sku := strings.TrimSpace(*req.SKU)
		if sku == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "sku cannot be empty"})
		}
		listing.SKU = sku
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "name cannot be empty"})
		}
		listing.Name = name
	}
	if req.IncludedMiles != nil {
		listing.IncludedMiles = *req.IncludedMiles
	}
	if req.Deposit != nil {
		listing.Deposit = *req.Deposit
	}
	if req.Available != nil {
		listing.Available = *req.Available
	}
	h.Store.SaveListing(listing)
	h.Logger.Info("admin_listing_updated", "listingID", listing.ID)
	return c.JSON(http.StatusOK, listing)
}

func (h *Handler) AdminDeleteListing(c echo.Context) error {
	listingID := c.Param("listingID")
	if _, ok := h.Store.GetListing(listingID); !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "listing not found"})
	}
	h.Store.DeleteListing(listingID)
	h.Logger.Info("admin_listing_deleted", "listingID", listingID)
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AdminBulkListings(c echo.Context) error {
	var req struct {
		ListingIDs    []string `json:"listingIds"`
		Available     *bool    `json:"available"`
		IncludedMiles *float64 `json:"includedMiles"`
		Deposit       *float64 `json:"deposit"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}
	if len(req.ListingIDs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "listingIds are required"})
	}
	if req.Available == nil && req.IncludedMiles == nil && req.Deposit == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "at least one field to update is required"})
	}

	updated := make([]models.Listing, 0, len(req.ListingIDs))
	for _, id := range req.ListingIDs {
		listing, ok := h.Store.GetListing(strings.TrimSpace(id))
		if !ok {
			continue
		}
		if req.Available != nil {
			listing.Available = *req.Available
		}
		if req.IncludedMiles != nil {
			listing.IncludedMiles = *req.IncludedMiles
		}
		if req.Deposit != nil {
			listing.Deposit = *req.Deposit
		}
		h.Store.SaveListing(listing)
		updated = append(updated, listing)
	}
	h.Logger.Info("admin_listing_bulk_update", "updatedCount", len(updated))
	return c.JSON(http.StatusOK, map[string]interface{}{
		"updatedCount": len(updated),
		"listings":     updated,
	})
}

func (h *Handler) AdminSearchListings(c echo.Context) error {
	q := strings.ToLower(strings.TrimSpace(c.QueryParam("q")))
	if q == "" {
		return c.JSON(http.StatusOK, h.Store.ListListings())
	}
	all := h.Store.ListListings()
	out := make([]models.Listing, 0, len(all))
	for _, listing := range all {
		if strings.Contains(strings.ToLower(listing.Name), q) || strings.Contains(strings.ToLower(listing.SPU), q) || strings.Contains(strings.ToLower(listing.SKU), q) {
			out = append(out, listing)
		}
	}
	return c.JSON(http.StatusOK, out)
}
