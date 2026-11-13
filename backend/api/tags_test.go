package api

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
	"note-taker/backend/models"
)

func TestTagShareTokenSchema(t *testing.T) {
	s, err := schema.Parse(&models.Tag{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("failed to parse Tag schema: %v", err)
	}
	f := s.FieldsByName["ShareToken"]
	if f == nil {
		t.Fatalf("field ShareToken not found on Tag")
	}
	if f.DBName != "share_token" {
		t.Errorf("expected DBName 'share_token', got '%s'", f.DBName)
	}
}

func TestBuildSubtreeTagPathStopsAtRoot(t *testing.T) {
	rootID := uint(1)
	sharedID := uint(2)
	childID := uint(3)
	grandchildID := uint(4)

	byID := map[uint]models.Tag{
		rootID:       {ID: rootID, ParentID: nil, Name: "PrivateRoot"},
		sharedID:     {ID: sharedID, ParentID: &rootID, Name: "SharedProject"},
		childID:      {ID: childID, ParentID: &sharedID, Name: "Sprint1"},
		grandchildID: {ID: grandchildID, ParentID: &childID, Name: "Standups"},
	}

	got := buildSubtreeTagPath(grandchildID, sharedID, byID)
	want := "SharedProject / Sprint1 / Standups"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}

	gotRoot := buildSubtreeTagPath(sharedID, sharedID, byID)
	if gotRoot != "SharedProject" {
		t.Errorf("expected %q, got %q", "SharedProject", gotRoot)
	}
}
