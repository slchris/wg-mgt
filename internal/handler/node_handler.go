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

// NodeHandler handles node HTTP requests.
type NodeHandler struct {
	nodeService *service.NodeService
}

// NewNodeHandler creates a new NodeHandler.
func NewNodeHandler(nodeService *service.NodeService) *NodeHandler {
	return &NodeHandler{nodeService: nodeService}
}

// CreateNodeRequest represents a request to create a node.
type CreateNodeRequest struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	SSHPort     int    `json:"ssh_port"`
	SSHUser     string `json:"ssh_user"`
	SSHKey      string `json:"ssh_key"`
	WGInterface string `json:"wg_interface"`
	WGPort      int    `json:"wg_port"`
	WGAddress   string `json:"wg_address"`
	Endpoint    string `json:"endpoint"`
	NetworkID   *uint  `json:"network_id"`
}

// Create creates a new node.
func (h *NodeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	node := &domain.Node{
		Name:        req.Name,
		Host:        req.Host,
		SSHPort:     req.SSHPort,
		SSHUser:     req.SSHUser,
		SSHKey:      req.SSHKey,
		WGInterface: req.WGInterface,
		WGPort:      req.WGPort,
		WGAddress:   req.WGAddress,
		Endpoint:    req.Endpoint,
		NetworkID:   req.NetworkID,
	}

	if err := h.nodeService.Create(node); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Created(w, node)
}

// GetAll retrieves all nodes.
func (h *NodeHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.nodeService.GetAll()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, nodes)
}

// GetByID retrieves a node by ID.
func (h *NodeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	node, err := h.nodeService.GetByID(uint(id))
	if err != nil {
		response.NotFound(w, "node not found")
		return
	}
	response.Success(w, node)
}

// Update updates a node.
func (h *NodeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	var req CreateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	node, err := h.nodeService.GetByID(uint(id))
	if err != nil {
		response.NotFound(w, "node not found")
		return
	}

	node.Name = req.Name
	node.Host = req.Host
	node.SSHPort = req.SSHPort
	node.SSHUser = req.SSHUser
	if req.SSHKey != "" {
		node.SSHKey = req.SSHKey
	}
	node.WGInterface = req.WGInterface
	node.WGPort = req.WGPort
	node.WGAddress = req.WGAddress
	node.Endpoint = req.Endpoint
	node.NetworkID = req.NetworkID

	if err := h.nodeService.Update(node); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, node)
}

// Delete deletes a node.
func (h *NodeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	if err := h.nodeService.Delete(uint(id)); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// CheckStatus checks the connectivity status of a node.
func (h *NodeHandler) CheckStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	node, err := h.nodeService.CheckStatus(uint(id))
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, node)
}

// GetWireGuardStatus retrieves WireGuard status from a node via SSH.
func (h *NodeHandler) GetWireGuardStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	status, err := h.nodeService.GetWireGuardStatus(uint(id))
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, status)
}

// GetSystemInfo retrieves system information from a node via SSH.
func (h *NodeHandler) GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	info, err := h.nodeService.GetNodeSystemInfo(uint(id))
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, info)
}
