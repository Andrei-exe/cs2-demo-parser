package jobs

import (
	"cs2-parser/internal/models"
	"time"
)

type Status string

const (
	StatusQueued  Status = "queued"
	StatusRunning Status = "running"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

type Job struct {
	ID         string             `json:"id"`
	Status     Status             `json:"status"`
	Progress   float64            `json:"progress"`
	Error      string             `json:"error,omitempty"`
	Result     *models.MatchStats `json:"result,omitempty"`
	CreatedAt  time.Time          `json:"createdAt"`
	FinishedAt time.Time          `json:"finishedAt,omitzero"`
	FilePath   string             `json:"-"`
}
