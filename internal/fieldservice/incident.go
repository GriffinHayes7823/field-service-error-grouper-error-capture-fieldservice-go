package fieldservice

import (
	"errors"
	"fmt"
)

type Area string

const (
	PhotoProcessing Area = "photo_processing"
	DispatchStatus  Area = "dispatch_status"
	TechFollowUp    Area = "technician_follow_up"
)

type Incident struct {
	WorkOrderID string `json:"work_order_id"`
	Area        Area   `json:"area"`
	PhotoID     string `json:"photo_id,omitempty"`
	Technician  string `json:"technician_id,omitempty"`
	Message     string `json:"message"`
	Exception   string `json:"exception"`
}

type Capture struct {
	Title          string         `json:"title"`
	Message        string         `json:"message"`
	Level          string         `json:"level"`
	Fingerprint    []string       `json:"fingerprint"`
	Exception      string         `json:"exception"`
	Context        map[string]any `json:"context"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
}

func (i Incident) Capture() (Capture, error) {
	if i.WorkOrderID == "" || i.Message == "" || i.Exception == "" {
		return Capture{}, errors.New("work_order_id, message, and exception are required")
	}

	fingerprint := []string{"work-order", i.WorkOrderID, string(i.Area)}
	context := map[string]any{"work_order_id": i.WorkOrderID, "area": i.Area}

	switch i.Area {
	case PhotoProcessing:
		if i.PhotoID == "" {
			return Capture{}, errors.New("photo_id is required for photo_processing")
		}
		context["photo_id"] = i.PhotoID
	case DispatchStatus:
		// Dispatch failures group at the work-order stage, independent of assignee.
	case TechFollowUp:
		if i.Technician == "" {
			return Capture{}, errors.New("technician_id is required for technician_follow_up")
		}
		fingerprint = append(fingerprint, i.Technician)
		context["technician_id"] = i.Technician
	default:
		return Capture{}, fmt.Errorf("unknown area %q", i.Area)
	}

	return Capture{
		Title:       fmt.Sprintf("work order %s: %s", i.WorkOrderID, i.Area),
		Message:     i.Message,
		Level:       "error",
		Fingerprint: fingerprint,
		Exception:   i.Exception,
		Context:     context,
	}, nil
}
