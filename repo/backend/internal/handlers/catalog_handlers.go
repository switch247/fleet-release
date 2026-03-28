package handlers

import (
	"net/http"
	"strings"

	"fleetlease/backend/internal/models"

	"github.com/labstack/echo/v4"
)

type categoryNode struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	ParentID string          `json:"parentId,omitempty"`
	Children []*categoryNode `json:"children,omitempty"`
}

func (h *Handler) Categories(c echo.Context) error {
	categories := h.Store.ListCategories()
	if strings.EqualFold(strings.TrimSpace(c.QueryParam("view")), "tree") {
		return c.JSON(http.StatusOK, buildCategoryTree(categories))
	}
	return c.JSON(http.StatusOK, categories)
}

func (h *Handler) StatsSummary(c echo.Context) error {
	bookings := h.Store.ListBookings()
	out := models.StatsSummary{}
	for _, b := range bookings {
		if b.Status == "settled" {
			out.SettledTrips++
		} else {
			out.ActiveBookings++
			out.HeldDeposits += b.DepositAmount
		}
	}
	out.InspectionsDue = out.ActiveBookings
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) Listings(c echo.Context) error {
	return c.JSON(http.StatusOK, h.Store.ListListings())
}

func buildCategoryTree(categories []models.Category) []*categoryNode {
	nodes := make(map[string]*categoryNode, len(categories))
	for _, category := range categories {
		item := &categoryNode{
			ID:       category.ID,
			Name:     category.Name,
			ParentID: category.ParentID,
		}
		nodes[category.ID] = item
	}

	roots := make([]*categoryNode, 0)
	for _, category := range categories {
		node, ok := nodes[category.ID]
		if !ok {
			continue
		}
		if category.ParentID != "" {
			if parent, parentOK := nodes[category.ParentID]; parentOK {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}
	return roots
}
