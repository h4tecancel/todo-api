package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"todo-api/internal/storage/sqlite"

	"github.com/gorilla/mux"
)

const (
	// JSON / запрос
	msgInvalidJSON    = "invalid JSON"
	msgInvalidRequest = "invalid request body"
	msgInvalidID      = "invalid id"

	// База данных
	msgDBError     = "database error"
	msgSaveError   = "could not save task"
	msgUpdateError = "could not update task"
	msgDeleteError = "could not delete task"
	msgSelectError = "could not fetch task"
)

type Handlers struct {
	Logger *slog.Logger
	DB     *sqlite.Storage
}

func New(log *slog.Logger, db *sqlite.Storage) *Handlers {
	return &Handlers{
		Logger: log,
		DB:     db,
	}
}

func (h *Handlers) AddNewTask(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.add_new_task"

	ctx := r.Context()
	defer r.Body.Close()

	var task TaskDTO
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&task); err != nil {
		WriteJSONError(w, h.Logger, op, http.StatusBadRequest, msgInvalidJSON, err)
		return
	}

	if err := task.ValidateForCreate(); err != nil {

		WriteJSONError(w, h.Logger, op, http.StatusBadRequest, "field validation failed", err)
		return
	}

	id, err := h.DB.SaveTask(ctx, task.Name, task.Description)
	if err != nil {
		WriteJSONError(w, h.Logger, op, http.StatusInternalServerError, msgSaveError, err)
		return
	}

	resp := TaskResponse{
		ID:        id,
		Name:      task.Name,
		Execution: "OK",
	}

	w.Header().Set("Location", fmt.Sprintf("/tasks/%d", id))
	WriteJSON(w, http.StatusCreated, resp)

	h.Logger.Info("task created",
		slog.String("op", op),
		slog.Int64("id", id),
	)
}

func (h *Handlers) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.get_task_by_id"

	ctx := r.Context()
	defer r.Body.Close()

	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		WriteJSONError(w, h.Logger, op, http.StatusBadRequest, msgInvalidID, err)
		return
	}

	task, err := h.DB.GetTaskByID(ctx, id)
	if err != nil {
		WriteJSONError(w, h.Logger, op, http.StatusInternalServerError, msgSelectError, err)
		return
	}

	WriteJSON(w, http.StatusOK, task)
	h.Logger.Info("get task",
		slog.String("op", op),
		slog.Int64("id", id),
	)
}

func (h *Handlers) CompleteTask(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.complete_task"

	ctx := r.Context()
	defer r.Body.Close()

	var req CompleteTaskDTO
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		WriteJSONError(w, h.Logger, op, http.StatusBadRequest, msgInvalidJSON, err)
		return
	}

	// базовая валидация
	if req.ID == 0 {
		WriteJSONError(w, h.Logger, op, http.StatusBadRequest, msgInvalidID, nil)
		return
	}
	if !req.Complete {
		WriteJSONError(w, h.Logger, op, http.StatusBadRequest, "complete must be true", nil)
		return
	}

	task, err := h.DB.GetTaskByID(ctx, req.ID)
	if err != nil {
		WriteJSONError(w, h.Logger, op, http.StatusInternalServerError, "failed to load task before update", err)
		return
	}

	if err := h.DB.CompleteByID(ctx, req.ID); err != nil {
		WriteJSONError(w, h.Logger, op, http.StatusInternalServerError, msgUpdateError, err)
		return
	}

	resp := TaskResponse{
		ID:        req.ID,
		Name:      task.Name,
		Execution: "complete status is OK",
	}
	WriteJSON(w, http.StatusOK, resp)

	h.Logger.Info("complete task",
		slog.String("op", op),
		slog.Int64("id", req.ID),
	)
}


func (h *Handlers) GetTasks(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.get_tasks"

	ctx := r.Context()
	defer r.Body.Close()

	query := r.URL.Query().Get("complete")
	switch query {
	case "true":
		tasks, err := h.DB.GetCompleteTasks(ctx)
		if err != nil {
			WriteJSONError(w, h.Logger, op, http.StatusInternalServerError, msgDBError, err)
			return
		}
		WriteJSON(w, http.StatusOK, tasks)
		return
	case "false":
		tasks, err := h.DB.GetNotCompleteTasks(ctx)
		if err != nil {
			WriteJSONError(w, h.Logger, op, http.StatusInternalServerError, msgDBError, err)
			return
		}
		WriteJSON(w, http.StatusOK, tasks)
		return
	default:
		tasks, err := h.DB.GetAllTasks(ctx)
		if err != nil {
			WriteJSONError(w, h.Logger, op, http.StatusInternalServerError, msgDBError, err)
			return
		}
		WriteJSON(w, http.StatusOK, tasks)
		h.Logger.Info("get all tasks", slog.String("op", op))
		return
	}
}


func (h *Handlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.delete_task"

	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		WriteJSONError(w, h.Logger, op, http.StatusBadRequest, msgInvalidID, err)
		return
	}

	ctx := r.Context()
	defer r.Body.Close()

	task, err := h.DB.GetTaskByID(ctx, id)
	if err != nil {
		WriteJSONError(w, h.Logger, op, http.StatusInternalServerError, msgSelectError, err)
		return
	}

	if err := h.DB.DeleteTask(ctx, id); err != nil {
		WriteJSONError(w, h.Logger, op, http.StatusInternalServerError, msgDeleteError, err)
		return
	}

	resp := TaskResponse{
		ID:        id,
		Name:      task.Name,
		Execution: "delete status is OK",
	}
	WriteJSON(w, http.StatusOK, resp)

	h.Logger.Info("task deleted",
		slog.String("op", op),
		slog.Int64("task_id", id),
	)
}
