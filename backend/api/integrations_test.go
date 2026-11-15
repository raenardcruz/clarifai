package api

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
	"note-taker/backend/models"
)

func TestRecordingAndSettingsIntegrationSchema(t *testing.T) {
	recSchema, err := schema.Parse(&models.Recording{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("failed to parse Recording schema: %v", err)
	}
	for _, fieldName := range []string{"Source", "ExternalID", "ExternalURL"} {
		if recSchema.FieldsByName[fieldName] == nil {
			t.Errorf("expected Recording field %s", fieldName)
		}
	}

	setSchema, err := schema.Parse(&models.Settings{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("failed to parse Settings schema: %v", err)
	}
	expectedFields := []string{
		"FathomEnabled", "FathomMethod", "FathomAPIKey", "FathomMCPURL", "FathomLastSyncAt", "FathomLastSyncStatus",
		"FirefliesEnabled", "FirefliesMethod", "FirefliesAPIKey", "FirefliesMCPURL", "FirefliesLastSyncAt", "FirefliesLastSyncStatus",
		"OtterEnabled", "OtterMethod", "OtterAPIKey", "OtterAPIBaseURL", "OtterMCPURL", "OtterLastSyncAt", "OtterLastSyncStatus",
	}
	for _, fieldName := range expectedFields {
		if setSchema.FieldsByName[fieldName] == nil {
			t.Errorf("expected Settings field %s", fieldName)
		}
	}
}

func TestMaskSecretAndApplyMaskedSecretUpdate(t *testing.T) {
	secret := "real-api-key"
	target := &secret

	// Sending masked placeholder keeps existing secret untouched
	masked := "••••••••"
	applyMaskedSecretUpdate(&target, &masked)
	if target == nil || *target != "real-api-key" {
		t.Fatalf("expected existing secret to be preserved, got %v", target)
	}

	// Sending new key updates secret
	updated := "new-api-key"
	applyMaskedSecretUpdate(&target, &updated)
	if target == nil || *target != "new-api-key" {
		t.Fatalf("expected secret to update, got %v", target)
	}

	// Sending empty clears secret
	empty := ""
	applyMaskedSecretUpdate(&target, &empty)
	if target != nil {
		t.Fatalf("expected secret to be cleared, got %v", *target)
	}

	// buildSettingsOut masks keys and fills defaults
	s := models.Settings{
		FathomEnabled:    true,
		FathomAPIKey:     &updated,
		FirefliesEnabled: false,
	}
	out := buildSettingsOut(&s)
	if out.FathomAPIKey == nil || *out.FathomAPIKey != "••••••••" {
		t.Errorf("expected masked FathomAPIKey, got %v", out.FathomAPIKey)
	}
	if out.FathomMethod != "api" || out.FirefliesMethod != "api" || out.OtterMethod != "api" {
		t.Errorf("expected default 'api' methods, got %s/%s/%s", out.FathomMethod, out.FirefliesMethod, out.OtterMethod)
	}
	if out.OtterAPIBaseURL != "https://api.otter.ai" {
		t.Errorf("expected default OtterAPIBaseURL, got %q", out.OtterAPIBaseURL)
	}
}

func TestDeletedImportSchema(t *testing.T) {
	delSchema, err := schema.Parse(&models.DeletedImport{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("failed to parse DeletedImport schema: %v", err)
	}
	if delSchema.Table != "deleted_imports" {
		t.Errorf("expected table 'deleted_imports', got %q", delSchema.Table)
	}
	for _, fieldName := range []string{"UserID", "Source", "ExternalID", "DeletedAt"} {
		if delSchema.FieldsByName[fieldName] == nil {
			t.Errorf("expected DeletedImport field %s", fieldName)
		}
	}
	idx := delSchema.LookIndex("idx_deleted_import_user_source_ext")
	if idx == nil || idx.Class != "UNIQUE" || len(idx.Fields) != 3 {
		t.Fatalf("expected unique composite index idx_deleted_import_user_source_ext on 3 fields, got %+v", idx)
	}
}
