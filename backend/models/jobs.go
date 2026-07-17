package models

import (
	"sync"
)

type JobStatus struct {
	ID           string `json:"id"`
	Filename     string `json:"filename"`
	Status       string `json:"status"` // pending, transcribing, diarizing, summarizing, completed, error
	Progress     int    `json:"progress"`
	TranscriptMD string `json:"transcript_md,omitempty"`
	SummaryMD    string `json:"summary_md,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type JobsDB struct {
	mu   sync.RWMutex
	jobs map[string]JobStatus
}

var GlobalJobsDB = &JobsDB{
	jobs: make(map[string]JobStatus),
}

func (db *JobsDB) Set(id string, status JobStatus) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.jobs[id] = status
}

func (db *JobsDB) Get(id string) (JobStatus, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	status, exists := db.jobs[id]
	return status, exists
}

func (db *JobsDB) Delete(id string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.jobs, id)
}
