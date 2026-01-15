package main

import (
	"time"

	"github.com/UkaSystems/chirpy/internal/database"
	"github.com/google/uuid"
)

type ChirpResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func ChirpResponseFromDB(cm *database.Chirp) ChirpResponse {
	return ChirpResponse{
		Id:        cm.ID,
		CreatedAt: cm.CreatedAt,
		UpdatedAt: cm.UpdatedAt,
		Body:      cm.Body,
		UserId:    cm.UserID,
	}
}
