package models

import (
	"time"
)

type User struct {
	ID                      uint        `gorm:"primaryKey" json:"id"`
	Email                   string      `gorm:"uniqueIndex;not null" json:"email"`
	HashedPassword          string      `gorm:"not null" json:"-"`
	Role                    string      `gorm:"default:user" json:"role"`
	IsApproved              bool        `gorm:"default:false" json:"is_approved"`
	TranscriptionLimitHours float64     `gorm:"default:40.0;not null" json:"transcription_limit_hours"`
	CreatedAt               time.Time   `gorm:"autoCreateTime" json:"created_at"`
	Recordings              []Recording `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"recordings,omitempty"`
}

type Recording struct {
	ID                string              `gorm:"primaryKey" json:"id"`
	UserID            uint                `json:"user_id"`
	Title             string              `gorm:"not null" json:"title"`
	Filename          string              `gorm:"not null" json:"filename"`
	AudioPath         string              `gorm:"not null" json:"audio_path"`
	Status            string              `gorm:"default:pending" json:"status"` // pending, processing, completed, error
	SummaryMD         string              `json:"summary_md,omitempty"`
	Duration          *float64            `json:"duration,omitempty"`
	SpeechmaticsJobID *string             `json:"speechmatics_job_id,omitempty"`
	ErrorMessage      *string             `json:"error_message,omitempty"`
	CreatedAt         time.Time           `gorm:"autoCreateTime" json:"created_at"`
	Segments          []TranscriptSegment `gorm:"foreignKey:RecordingID;constraint:OnDelete:CASCADE" json:"segments,omitempty"`
}

type TranscriptSegment struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	RecordingID string  `gorm:"index;not null" json:"recording_id"`
	StartTime   float64 `gorm:"not null" json:"start_time"`
	EndTime     float64 `gorm:"not null" json:"end_time"`
	Speaker     string  `gorm:"not null" json:"speaker"`
	Text        string  `gorm:"not null" json:"text"`
}

type Settings struct {
	ID                     uint    `gorm:"primaryKey" json:"id"`
	AISummarizationMode    string  `gorm:"default:auto" json:"ai_summarization_mode"`
	ExecutiveSummaryPrompt string  `gorm:"type:text;default:'Analyze the following transcript and write a detailed, cohesive executive summary in a narrative paragraph format. Synthesize the meeting''s core purpose, main arguments, and final outcomes into smooth, professional prose, explicitly attributing key ideas, decisions, and viewpoints to specific speakers by name (or speaker identifier) directly within the flow of the text. Ensure the summary is comprehensive and captures concrete details, specific project names, and actionable next steps, but deliver it entirely as a sequence of well-structured paragraphs without using any bullet points, lists, or tables.'"`
	SpeechmaticsAPIKey     *string `json:"speechmatics_api_key,omitempty"`
	OllamaModel            string  `gorm:"default:'gemma4:12b-mlx'" json:"ollama_model"`
}
