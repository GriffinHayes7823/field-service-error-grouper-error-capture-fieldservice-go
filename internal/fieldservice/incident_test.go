package fieldservice

import (
	"reflect"
	"testing"
)

func TestIncidentCaptureGrouping(t *testing.T) {
	tests := []struct {
		name        string
		incident    Incident
		fingerprint []string
		wantErr     bool
	}{
		{
			name: "photos from one stage group despite different photo ids",
			incident: Incident{WorkOrderID: "WO-42", Area: PhotoProcessing, PhotoID: "IMG-9",
				Message: "thumbnail generation failed", Exception: "decode: invalid image data"},
			fingerprint: []string{"work-order", "WO-42", "photo_processing"},
		},
		{
			name: "dispatch groups by work order and stage",
			incident: Incident{WorkOrderID: "WO-42", Area: DispatchStatus,
				Message: "status transition rejected", Exception: "assigned cannot move to completed"},
			fingerprint: []string{"work-order", "WO-42", "dispatch_status"},
		},
		{
			name: "follow-up keeps technicians separate",
			incident: Incident{WorkOrderID: "WO-42", Area: TechFollowUp, Technician: "TECH-7",
				Message: "follow-up note failed", Exception: "note persistence failed"},
			fingerprint: []string{"work-order", "WO-42", "technician_follow_up", "TECH-7"},
		},
		{
			name: "photo requires an id",
			incident: Incident{WorkOrderID: "WO-42", Area: PhotoProcessing,
				Message: "decode failed", Exception: "invalid image data"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.incident.Capture()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Capture() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got.Fingerprint, tt.fingerprint) {
				t.Fatalf("fingerprint = %#v, want %#v", got.Fingerprint, tt.fingerprint)
			}
		})
	}
}
