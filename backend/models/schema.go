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
	IsMedical         bool                `gorm:"default:false" json:"is_medical"`
	CreatedAt         time.Time           `gorm:"autoCreateTime" json:"created_at"`
	Segments          []TranscriptSegment `gorm:"foreignKey:RecordingID;constraint:OnDelete:CASCADE" json:"segments,omitempty"`
	Tags              []Tag               `gorm:"many2many:recording_tags;constraint:OnDelete:CASCADE" json:"tags,omitempty"`
}

type Tag struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	UserID     uint        `gorm:"index;not null" json:"user_id"`
	ParentID   *uint       `gorm:"index" json:"parent_id"`
	Name       string      `gorm:"not null" json:"name"`
	Color      string      `gorm:"default:'#3b82f6';not null" json:"color"`
	ShareToken *string     `gorm:"uniqueIndex" json:"share_token,omitempty"`
	CreatedAt  time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
	Children   []Tag       `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Recordings []Recording `gorm:"many2many:recording_tags;constraint:OnDelete:CASCADE" json:"-"`
}

type TranscriptSegment struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	RecordingID string  `gorm:"index;not null" json:"recording_id"`
	StartTime   float64 `gorm:"not null" json:"start_time"`
	EndTime     float64 `gorm:"not null" json:"end_time"`
	Speaker     string  `gorm:"not null" json:"speaker"`
	Text        string  `gorm:"not null" json:"text"`
}

type MeetingInvite struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Token       string    `gorm:"uniqueIndex;not null" json:"token"`
	RecordingID string    `gorm:"index;not null" json:"recording_id"`
	CreatedBy   uint      `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type RecordingAccess struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	RecordingID string    `gorm:"uniqueIndex:idx_rec_user;not null" json:"recording_id"`
	UserID      uint      `gorm:"uniqueIndex:idx_rec_user;not null" json:"user_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

const DefaultExecutiveSummaryPrompt = `Analyze the following transcript and produce a comprehensive, deeply structured session summary that takes full advantage of the HTML summary template's design system:
- **Executive Overview & KPIs**: Synthesize the core thesis, primary topics, speaker viewpoints, and final outcomes in a crisp 2-4 sentence lead paragraph. If the transcript contains 2-3 clear headline metrics, counts, targets, or durations, populate the .kpi-grid cards.
- **Chronological Flow & Conceptual Breakdown**: Organize into sequential <h3> subsections (2.1, 2.2, ...) matching the exact chronological progression of the session. Attribute key ideas and decisions to specific speakers. Use <ol>/<ul> for step-by-step mechanisms, .callout boxes for concrete examples/analogies/case studies, <blockquote> for notable verbatim quotes, and <pre><code> for technical syntax, formulas, or commands when applicable.
- **Visual Analytics & Benchmarks**: Whenever the transcript includes numerical breakdowns, percentages, time allocations, scores, or trends, create proportional CSS horizontal bar charts (.bar-item) and/or inline SVG trend line charts (or clean inline SVG diagrams) in Section 03.
- **Comparative Analysis & Action Matrix**: Whenever 2+ options, approaches, tools, or phases are compared, build a structured comparison <table>. Extract all concrete tasks, decisions, and next steps into the Follow-Up Action Items table with owners, due dates, and status pills (.pill-done, .pill-progress, .pill-blocked).
- **Questions & Answers**: Capture audience or participant questions and their complete answers in .qa-card blocks.
- **Key Terms & Glossary**: Extract domain-specific terms, acronyms, and key concepts into .glossary-item tags with clear definitions.
- Cleanly omit any optional section or component that has no supporting data in the transcript and renumber remaining .sec-num badges sequentially.`

const DefaultMedicalSummaryPrompt = `Analyze the following medical or clinical encounter transcript and generate a structured Clinical / SOAP Note using the HTML summary template's full layout:
- **Header & Overview (Section 01)**: Set the eyebrow to "Clinical Consultation — SOAP Note" and summarize the Chief Complaint and reason for consultation in the lead paragraph. Include .kpi-grid cards ONLY if concrete vitals, lab values, or pain/severity scores are explicitly stated.
- **Chronological Clinical Breakdown (Section 02)**: Organize into <h3> subsections:
  - 2.1 Subjective (History of Present Illness): patient-reported symptoms, timeline, pain/severity, and relevant medical history.
  - 2.2 Objective & Clinical Discussion: observed clinical signs, vital signs, lab/diagnostic results mentioned, and physical exam discussions.
  - 2.3 Assessment & Clinical Impression: differential or confirmed diagnoses, clinical reasoning, and provider impressions.
  - 2.4 Plan & Management: prescriptions/medications (with dosage and instructions if discussed), ordered tests/referrals, lifestyle advice, and follow-up plan.
  Use .callout boxes for critical clinical warnings, allergies, or red flags, and <blockquote> for key verbatim patient statements when helpful.
- **Visual Analytics (Section 03)**: Include ONLY if numerical vitals, lab trends, or symptom/dosage comparisons are discussed.
- **Structured Clinical Tables (Section 04)**: Use a comparison <table> if multiple diagnostic or treatment options are weighed, and use the Action Matrix table for medications, ordered tests, referrals, and follow-up items with status pills (.pill-done, .pill-progress, .pill-blocked).
- **Patient Q&A (Section 05)** and **Clinical Glossary (Section 06)**: Capture patient questions and clinician explanations in .qa-card blocks, and define medical terms/medications in .glossary-item tags.
Ensure strict clinical accuracy without inventing data; cleanly omit any empty optional sections.`

type Settings struct {
	ID                     uint    `gorm:"primaryKey" json:"id"`
	AISummarizationMode    string  `gorm:"column:ai_summarization_mode;default:auto" json:"ai_summarization_mode"`
	ExecutiveSummaryPrompt string  `gorm:"column:executive_summary_prompt;type:text" json:"executive_summary_prompt"`
	MedicalSummaryPrompt   string  `gorm:"column:medical_summary_prompt;type:text" json:"medical_summary_prompt"`
	SpeechmaticsAPIKey     *string `gorm:"column:speechmatics_api_key" json:"speechmatics_api_key,omitempty"`
	OllamaURL              string  `gorm:"column:ollama_url;default:'http://localhost:11434'" json:"ollama_url"`
	OllamaModel            string  `gorm:"column:ollama_model;default:'gemma4:12b-mlx'" json:"ollama_model"`
	AIProvider             string  `gorm:"column:ai_provider;default:'ollama'" json:"ai_provider"`
	GeminiModel            string  `gorm:"column:gemini_model;default:'gemini-3.8-flash'" json:"gemini_model"`
	GeminiAPIKey           *string `gorm:"column:gemini_api_key" json:"gemini_api_key,omitempty"`
	GeminiAPIBase          string  `gorm:"column:gemini_api_base" json:"gemini_api_base,omitempty"`
}


