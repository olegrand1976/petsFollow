package handlers

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func TestStripClientConsultationFlags_NonOwnerClearsHasReport(t *testing.T) {
	items := []store.TimelineItem{
		{
			ID:   "v1",
			Type: kernel.TimelineVisit,
			Meta: map[string]any{"hasReport": true, "reportStatus": "final", "visitId": "v1"},
		},
		{
			ID:   "hr1",
			Type: kernel.TimelineHeartRate,
			Meta: map[string]any{"bpm": 72},
		},
	}
	stripClientConsultationFlags(items, "owner-id", "grantee-id", kernel.RoleClient)
	if _, ok := items[0].Meta["hasReport"]; ok {
		t.Fatalf("visit hasReport must be omitted, got %#v", items[0].Meta["hasReport"])
	}
	if _, ok := items[0].Meta["reportStatus"]; ok {
		t.Fatalf("visit reportStatus must be omitted, got %#v", items[0].Meta["reportStatus"])
	}
	if items[0].Meta["visitId"] != "v1" {
		t.Fatalf("visitId must stay %#v", items[0].Meta)
	}
	if items[1].Meta["bpm"] != 72 {
		t.Fatalf("non-visit meta must stay intact %#v", items[1].Meta)
	}
}

func TestStripClientConsultationFlags_OwnerKeepsHasReport(t *testing.T) {
	items := []store.TimelineItem{
		{
			ID:   "v1",
			Type: kernel.TimelineVisit,
			Meta: map[string]any{"hasReport": true, "reportStatus": "final"},
		},
	}
	stripClientConsultationFlags(items, "owner-id", "owner-id", kernel.RoleClient)
	if items[0].Meta["hasReport"] != true {
		t.Fatalf("owner hasReport want true got %#v", items[0].Meta)
	}
}

func TestStripClientConsultationFlags_VetUnchanged(t *testing.T) {
	items := []store.TimelineItem{
		{
			ID:   "v1",
			Type: kernel.TimelineVisit,
			Meta: map[string]any{"hasReport": true, "reportStatus": "draft"},
		},
	}
	stripClientConsultationFlags(items, "owner-id", "vet-id", kernel.RoleVet)
	if items[0].Meta["hasReport"] != true || items[0].Meta["reportStatus"] != "draft" {
		t.Fatalf("vet meta must stay %#v", items[0].Meta)
	}
}

func TestStripClientConsultationFlags_NilMetaAndCarePro(t *testing.T) {
	items := []store.TimelineItem{
		{ID: "v1", Type: kernel.TimelineVisit, Meta: nil},
		{
			ID:   "v2",
			Type: kernel.TimelineVisit,
			Meta: map[string]any{"hasReport": true, "reportStatus": "final"},
		},
	}
	stripClientConsultationFlags(items, "owner-id", "care-id", kernel.RoleCarePro)
	if items[0].Meta != nil {
		t.Fatalf("nil meta must stay nil %#v", items[0].Meta)
	}
	if items[1].Meta["hasReport"] != true {
		t.Fatalf("care_pro must not strip %#v", items[1].Meta)
	}
}

func TestClearTimelineMeta(t *testing.T) {
	items := []store.TimelineItem{
		{
			ID:   "v1",
			Type: kernel.TimelineVisit,
			Meta: map[string]any{"hasReport": true, "visitId": "v1"},
		},
		{
			ID:   "hr1",
			Type: kernel.TimelineHeartRate,
			Meta: map[string]any{"bpm": 72},
		},
	}
	clearTimelineMeta(items)
	for i, it := range items {
		if it.Meta != nil {
			t.Fatalf("item[%d] Meta want nil got %#v", i, it.Meta)
		}
	}
}
