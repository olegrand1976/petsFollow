package handlers

import (
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

// stripClientConsultationFlags removes timeline meta.hasReport for non-owner clients.
// Aligns with ListVisits AttachFinalReportFlags (consultation CTA is owner-only):
// omit keys rather than soft-clear to false/"".
func stripClientConsultationFlags(items []store.TimelineItem, ownerUserID, viewerUserID string, role kernel.Role) {
	if role != kernel.RoleClient || ownerUserID == viewerUserID {
		return
	}
	for i := range items {
		if items[i].Type != kernel.TimelineVisit || items[i].Meta == nil {
			continue
		}
		delete(items[i].Meta, "hasReport")
		delete(items[i].Meta, "reportStatus")
	}
}

// clearTimelineMeta drops Meta on all timeline items (dossier PDF uses When/Title/Body only).
func clearTimelineMeta(items []store.TimelineItem) {
	for i := range items {
		items[i].Meta = nil
	}
}
