package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/slchris/wg-mgt/internal/domain"
	"github.com/slchris/wg-mgt/internal/pkg/response"
	"github.com/slchris/wg-mgt/internal/service"
)

// NetworkHandler handles network HTTP requests.
type NetworkHandler struct {
	networkService *service.NetworkService
}

// NewNetworkHandler creates a new NetworkHandler.
func NewNetworkHandler(networkService *service.NetworkService) *NetworkHandler {
	return &NetworkHandler{networkService: networkService}
}

// CreateNetworkRequest represents a request to create a network.
type CreateNetworkRequest struct {
	Name        string `json:"name"`
	CIDR        string `json:"cidr"`
	Gateway     string `json:"gateway"`
	Description string `json:"description"`
}

// Create creates a new network.
func (h *NetworkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	network := &domain.Network{
		Name:        req.Name,
		CIDR:        req.CIDR,
		Gateway:     req.Gateway,
		Description: req.Description,
	}

	if err := h.networkService.Create(network); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Created(w, network)
}

// GetAll retrieves all networks.
func (h *NetworkHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	networks, err := h.networkService.GetAll()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, networks)
}

// GetByID retrieves a network by ID.
func (h *NetworkHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	network, err := h.networkService.GetByID(uint(id))
	if err != nil {
		response.NotFound(w, "network not found")
		return
	}
	response.Success(w, network)
}

// Update updates a network.
func (h *NetworkHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	var req CreateNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	network, err := h.networkService.GetByID(uint(id))
	if err != nil {
		response.NotFound(w, "network not found")
		return
	}

	network.Name = req.Name
	network.CIDR = req.CIDR
	network.Gateway = req.Gateway
	network.Description = req.Description

	if err := h.networkService.Update(network); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, network)
}

// Delete deletes a network.
func (h *NetworkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	if err := h.networkService.Delete(uint(id)); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}
