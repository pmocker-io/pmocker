package service

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

func TestStateMachineValidFlow(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{})
	et := pmocker.PMEntityType{TypeCode: "x", ModuleCode: "m", Name: "X", Status: "draft"}
	if err := db.Create(&et).Error; err != nil {
		t.Fatal(err)
	}

	s := &StateMachineService{}
	if err := s.Transition(db, "pm_entity_types", et.ID, "draft", "reviewing"); err != nil {
		t.Fatalf("draft→reviewing: %v", err)
	}
	if err := s.Transition(db, "pm_entity_types", et.ID, "reviewing", "published"); err != nil {
		t.Fatalf("reviewing→published: %v", err)
	}
	if err := s.Transition(db, "pm_entity_types", et.ID, "published", "archived"); err != nil {
		t.Fatalf("published→archived: %v", err)
	}
	if err := s.Transition(db, "pm_entity_types", et.ID, "archived", "draft"); err != nil {
		t.Fatalf("archived→draft: %v", err)
	}
}

func TestStateMachineIllegalTransition(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{})
	et := pmocker.PMEntityType{TypeCode: "x", ModuleCode: "m", Name: "X", Status: "draft"}
	if err := db.Create(&et).Error; err != nil {
		t.Fatal(err)
	}

	s := &StateMachineService{}
	if err := s.Transition(db, "pm_entity_types", et.ID, "draft", "published"); err == nil {
		t.Fatal("draft→published 应被拒绝（必须经 reviewing）")
	}
}

func TestDeleteDraftOnly(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{})
	et := pmocker.PMEntityType{TypeCode: "x", ModuleCode: "m", Name: "X", Status: "published"}
	if err := db.Create(&et).Error; err != nil {
		t.Fatal(err)
	}

	s := &StateMachineService{}
	if err := s.DeleteDraft(db, "pm_entity_types", et.ID); err == nil {
		t.Fatal("published 状态不应被删除")
	}
}
