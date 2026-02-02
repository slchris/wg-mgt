package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/skip2/go-qrcode"
	"github.com/slchris/wg-mgt/internal/domain"
	"github.com/slchris/wg-mgt/internal/pkg/response"
	"github.com/slchris/wg-mgt/internal/service"
)

// PeerHandler handles peer HTTP requests.
type PeerHandler struct {
	peerService *service.PeerService
}

// NewPeerHandler creates a new PeerHandler.
func NewPeerHandler(peerService *service.PeerService) *PeerHandler {
	return &PeerHandler{peerService: peerService}
}

// CreatePeerRequest represents a request to create a peer.
type CreatePeerRequest struct {
	Name       string `json:"name"`
	NodeID     uint   `json:"node_id"`
	AllowedIPs string `json:"allowed_ips"`
	Address    string `json:"address"`
	DNS        string `json:"dns"`
	Enabled    bool   `json:"enabled"`
}

// Create creates a new peer.
func (h *PeerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreatePeerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	peer := &domain.Peer{
		Name:       req.Name,
		NodeID:     req.NodeID,
		AllowedIPs: req.AllowedIPs,
		Address:    req.Address,
		DNS:        req.DNS,
		Enabled:    req.Enabled,
	}

	if err := h.peerService.Create(peer); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Created(w, peer)
}

// GetAll retrieves all peers.
func (h *PeerHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	peers, err := h.peerService.GetAll()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, peers)
}

// GetByID retrieves a peer by ID.
func (h *PeerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	peer, err := h.peerService.GetByID(uint(id))
	if err != nil {
		response.NotFound(w, "peer not found")
		return
	}
	response.Success(w, peer)
}

// GetByNodeID retrieves all peers for a node.
func (h *PeerHandler) GetByNodeID(w http.ResponseWriter, r *http.Request) {
	nodeID, err := strconv.ParseUint(chi.URLParam(r, "nodeId"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid node id")
		return
	}

	peers, err := h.peerService.GetByNodeID(uint(nodeID))
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, peers)
}

// Update updates a peer.
func (h *PeerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	var req CreatePeerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	peer, err := h.peerService.GetByID(uint(id))
	if err != nil {
		response.NotFound(w, "peer not found")
		return
	}

	peer.Name = req.Name
	peer.AllowedIPs = req.AllowedIPs
	peer.Address = req.Address
	peer.DNS = req.DNS
	peer.Enabled = req.Enabled

	if err := h.peerService.Update(peer); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, peer)
}

// Delete deletes a peer.
func (h *PeerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	if err := h.peerService.Delete(uint(id)); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// GetConfig retrieves the WireGuard configuration for a peer.
func (h *PeerHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	config, err := h.peerService.GenerateClientConfig(uint(id))
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, map[string]string{"config": config})
}

// GetQRCode generates a QR code image for the peer's WireGuard configuration.
func (h *PeerHandler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	config, err := h.peerService.GenerateClientConfig(uint(id))
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	// Generate QR code PNG
	png, err := qrcode.Encode(config, qrcode.Medium, 256)
	if err != nil {
		response.InternalError(w, "failed to generate QR code")
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(png)
}

// GetNextIP returns the next available IP address for a node.
func (h *PeerHandler) GetNextIP(w http.ResponseWriter, r *http.Request) {
	nodeID, err := strconv.ParseUint(chi.URLParam(r, "nodeId"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid node id")
		return
	}

	nextIP, err := h.peerService.GetNextAvailableIP(uint(nodeID))
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, map[string]string{"next_ip": nextIP})
}

// SyncToServer syncs all peers for a node to the WireGuard server.
func (h *PeerHandler) SyncToServer(w http.ResponseWriter, r *http.Request) {
	nodeID, err := strconv.ParseUint(chi.URLParam(r, "nodeId"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid node id")
		return
	}

	if err := h.peerService.SyncAllPeersToServer(uint(nodeID)); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, map[string]string{"message": "peers synced successfully"})
}
