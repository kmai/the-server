package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kmai/the-server/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Handler struct {
	Database *gorm.DB
	Logger   *logrus.Logger
}

// GetUserList returns a list of Users.
func (h *Handler) GetUserList(writer http.ResponseWriter, _ *http.Request) {
	var results []models.User

	h.Database.Find(&results)

	serializedResults, err := json.Marshal(results)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		if _, err := writer.Write(serializedResults); err != nil {
			h.Logger.Error(fmt.Errorf("error while writing response: %w", err))
		}
	}
}

// GetUser returns a User.
func (h *Handler) GetUser(writer http.ResponseWriter, req *http.Request) {
	userID := chi.URLParam(req, "userID")

	result := models.User{ID: userID}

	h.Database.First(&result)

	serializedResults, err := json.Marshal(result)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		if _, err := writer.Write(serializedResults); err != nil {
			h.Logger.Error(fmt.Errorf("error while writing response: %w", err))
		}
	}
}

// PostUser creates a User.
func (h *Handler) PostUser(writer http.ResponseWriter, _ *http.Request) {
	user := &models.User{ID: uuid.NewString()}

	result := h.Database.Create(user)

	var statusHeader int

	var encodedPayload []byte

	var err error

	// Determine status code and payload
	if result.Error != nil {
		statusHeader = http.StatusBadRequest
		encodedPayload, err = json.Marshal(&map[string]string{
			"error": result.Error.Error(),
		})
	} else {
		statusHeader = http.StatusCreated
		encodedPayload, err = json.Marshal(user)
	}

	if err != nil {
		statusHeader = http.StatusInternalServerError

		encodedPayload, err = json.Marshal(&map[string]string{
			"error": err.Error(),
		})
		if err != nil {
			h.Logger.Warnf("error preparing response: %v", err)
		}
	}

	writer.WriteHeader(statusHeader)

	if _, err = writer.Write(encodedPayload); err != nil {
		h.Logger.Error(fmt.Errorf("error while writing response: %w", err))
	}
}
