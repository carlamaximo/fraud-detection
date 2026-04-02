package validators

import (
	"testing"

	"fraud-platform/internal/domain"
)

func TestValidateBusinessRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		event   domain.Event
		wantErr bool
	}{
		{
			name: "ok_br_known_device",
			event: domain.Event{
				UserID: "u1", EventType: "login", IP: "1.1.1.1",
				Country: "BR", Device: "iphone",
			},
			wantErr: false,
		},
		{
			name: "reject_br_unknown_device",
			event: domain.Event{
				UserID: "u1", EventType: "login", IP: "1.1.1.1",
				Country: "BR", Device: "unknown",
			},
			wantErr: true,
		},
		{
			name: "ok_us_unknown_device",
			event: domain.Event{
				UserID: "u1", EventType: "login", IP: "1.1.1.1",
				Country: "US", Device: "unknown",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateBusinessRules(tt.event)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
