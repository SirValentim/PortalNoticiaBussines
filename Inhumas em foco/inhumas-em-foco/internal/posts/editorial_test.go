package posts

import (
	"testing"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/model"
)

func TestEditorialServiceCanEdit(t *testing.T) {
	svc := NewEditorialService(auth.NewService(nil))
	authorID := int64(10)
	post := &model.Post{AuthorID: &authorID, Status: model.StatusDraft}

	if !svc.CanEdit(&model.User{ID: 1, Role: model.RoleEditor}, post) {
		t.Fatal("editor should edit any post")
	}
	if !svc.CanEdit(&model.User{ID: 10, Role: model.RoleRedator}, post) {
		t.Fatal("redator should edit own draft")
	}
	if svc.CanEdit(&model.User{ID: 11, Role: model.RoleRedator}, post) {
		t.Fatal("redator should not edit another author's post")
	}
	post.Status = model.StatusPublished
	if svc.CanEdit(&model.User{ID: 10, Role: model.RoleRedator}, post) {
		t.Fatal("redator should not edit own published post")
	}
}

func TestEditorialServiceValidateStatusPermission(t *testing.T) {
	svc := NewEditorialService(auth.NewService(nil))

	if msg := svc.ValidateStatusPermission(&model.User{Role: model.RoleRedator}, model.StatusPublished); msg == "" {
		t.Fatal("redator publish should be blocked")
	}
	if msg := svc.ValidateStatusPermission(&model.User{Role: model.RoleRedator}, model.StatusApproved); msg == "" {
		t.Fatal("redator approve should be blocked")
	}
	if msg := svc.ValidateStatusPermission(&model.User{Role: model.RoleRevisor}, model.StatusApproved); msg != "" {
		t.Fatalf("revisor approve msg = %q", msg)
	}
	if msg := svc.ValidateStatusPermission(&model.User{Role: model.RoleEditor}, model.StatusPublished); msg != "" {
		t.Fatalf("editor publish msg = %q", msg)
	}
	if msg := svc.ValidateStatusPermission(&model.User{Role: model.RoleRevisor}, model.StatusArchived); msg == "" {
		t.Fatal("revisor archive should be blocked")
	}
}

func TestEditorialServiceValidateFormAndEditorialNotes(t *testing.T) {
	svc := NewEditorialService(auth.NewService(nil))

	if msg := svc.ValidateForm(&model.Post{Content: "texto", Status: model.StatusDraft}); msg != "Titulo e obrigatorio" {
		t.Fatalf("title validation = %q", msg)
	}
	if msg := svc.ValidateForm(&model.Post{Title: "Titulo", Content: "texto", Status: "bad"}); msg != "Status invalido" {
		t.Fatalf("status validation = %q", msg)
	}
	if msg := svc.ValidateRequiredEditorialNotes(&model.Post{Title: "Titulo"}, true); msg == "" {
		t.Fatal("required editorial notes should be enforced")
	}
	if msg := svc.ValidateRequiredEditorialNotes(&model.Post{EditorialNotes: "apuracao", EditorResponsible: "Editor"}, true); msg != "" {
		t.Fatalf("editorial notes msg = %q", msg)
	}
}

func TestEditorialServiceWorkflowPermissions(t *testing.T) {
	svc := NewEditorialService(auth.NewService(nil))
	authorID := int64(10)
	draft := &model.Post{AuthorID: &authorID, Status: model.StatusDraft}
	review := &model.Post{AuthorID: &authorID, Status: model.StatusReview}

	if !svc.CanSubmitReview(&model.User{ID: 10, Role: model.RoleRedator}, draft) {
		t.Fatal("redator should submit own draft to review")
	}
	if svc.CanSubmitReview(&model.User{ID: 11, Role: model.RoleRedator}, draft) {
		t.Fatal("redator should not submit another author draft")
	}
	if !svc.CanApprove(&model.User{ID: 20, Role: model.RoleRevisor}, review) {
		t.Fatal("revisor should approve review posts")
	}
	if svc.CanApprove(&model.User{ID: 10, Role: model.RoleRedator}, review) {
		t.Fatal("redator should not approve review posts")
	}
}
