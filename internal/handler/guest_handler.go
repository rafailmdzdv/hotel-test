package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"hotel-test/internal/domain"
	"hotel-test/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type GuestHandler struct {
	guestService *service.GuestService
}

func NewGuestHandler(guestService *service.GuestService) *GuestHandler {
	return &GuestHandler{guestService: guestService}
}

type ListResponse struct {
	Data   []domain.GuestRecord `json:"data"`
	Total  int                  `json:"total"`
	Limit  int                  `json:"limit"`
	Offset int                  `json:"offset"`
}

func (h *GuestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateGuestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, fmt.Errorf("%w: invalid JSON format", domain.ErrInvalidInput))
		return
	}

	record, err := h.guestService.Create(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, record)
}

func (h *GuestHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, fmt.Errorf("%w: invalid id parameter", domain.ErrInvalidInput))
		return
	}

	record, err := h.guestService.GetByID(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, record)
}

func (h *GuestHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	limit := 20
	if l := query.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	offset := 0
	if o := query.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	filter := domain.GuestFilter{
		RoomNumber: query.Get("room_number"),
		Phone:      query.Get("phone"),
		Limit:      limit,
		Offset:     offset,
	}

	records, total, err := h.guestService.List(r.Context(), filter)
	if err != nil {
		WriteError(w, err)
		return
	}

	if records == nil {
		records = []domain.GuestRecord{}
	}

	resp := ListResponse{
		Data:   records,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	WriteJSON(w, http.StatusOK, resp)
}

func (h *GuestHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, fmt.Errorf("%w: invalid id parameter", domain.ErrInvalidInput))
		return
	}

	var req domain.UpdateGuestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, fmt.Errorf("%w: invalid JSON format", domain.ErrInvalidInput))
		return
	}

	record, err := h.guestService.Update(r.Context(), id, req)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, record)
}

func (h *GuestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, fmt.Errorf("%w: invalid id parameter", domain.ErrInvalidInput))
		return
	}

	if err := h.guestService.Delete(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusNoContent, nil)
}
