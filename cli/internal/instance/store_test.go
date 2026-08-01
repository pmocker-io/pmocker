package instance

import (
	"testing"
	"time"
)

func TestStoreCRUD(t *testing.T) {
	store, err := NewStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	now := time.Now()
	inst := &Instance{
		ID: "test-001", Name: "test-instance",
		ImageDigest: "sha256:abc", ImageRef: "pmbok6:latest",
		Port: 8080, VolumeID: "vol-001",
		PID: 0, Status: "stopped", CreatedAt: now,
	}

	if err := store.Create(inst); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := store.GetByID("test-001")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "test-instance" {
		t.Errorf("name = %s, want test-instance", got.Name)
	}

	got2, err := store.GetByName("test-instance")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if got2.ID != "test-001" {
		t.Errorf("id = %s, want test-001", got2.ID)
	}

	got.Status = "running"
	got.PID = 12345
	if err := store.Update(got); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got3, _ := store.GetByID("test-001")
	if got3.Status != "running" || got3.PID != 12345 {
		t.Errorf("after update: status=%s pid=%d", got3.Status, got3.PID)
	}

	list, err := store.List(true)
	if err != nil || len(list) != 1 {
		t.Fatalf("List: %v len=%d", err, len(list))
	}

	list2, _ := store.List(false)
	if len(list2) != 1 {
		t.Errorf("List(running only): len=%d, want 1", len(list2))
	}

	if err := store.Delete("test-001"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = store.GetByID("test-001")
	if err == nil {
		t.Error("expected error after delete")
	}
}
