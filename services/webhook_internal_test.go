package services

import (
	"testing"
)

func TestStripTrackers(t *testing.T) {
	tests := []struct {
		name            string
		html            string
		expectedHTML    string
		expectedBlocked int
	}{
		{
			name:            "normal image",
			html:            `<img src="https://khrees.com/logo.png" width="200" height="100">`,
			expectedHTML:    `<img src="https://khrees.com/logo.png" width="200" height="100">`,
			expectedBlocked: 0,
		},
		{
			name:            "1x1 pixel via width/height attributes",
			html:            `<img src="https://example.com/pixel.png" width="1" height="1">`,
			expectedHTML:    ``,
			expectedBlocked: 1,
		},
		{
			name:            "1x1 pixel via style",
			html:            `<img src="https://example.com/style.png" style="width: 1px; height: 1px;">`,
			expectedHTML:    ``,
			expectedBlocked: 1,
		},
		{
			name:            "known tracker domain",
			html:            `<img src="https://google-analytics.com/collect?v=1">`,
			expectedHTML:    ``,
			expectedBlocked: 1,
		},
		{
			name:            "known tracker path suffix",
			html:            `<img src="https://example.com/wf/open?id=abc">`,
			expectedHTML:    ``,
			expectedBlocked: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHTML, gotBlocked := stripTrackers(tt.html)
			if gotBlocked != tt.expectedBlocked {
				t.Errorf("stripTrackers() gotBlocked = %d; want %d", gotBlocked, tt.expectedBlocked)
			}
			if tt.expectedBlocked > 0 && gotHTML != "" {
				t.Errorf("stripTrackers() gotHTML = %q; want empty", gotHTML)
			}
			if tt.expectedBlocked == 0 && gotHTML != tt.expectedHTML {
				t.Errorf("stripTrackers() gotHTML = %q; want %q", gotHTML, tt.expectedHTML)
			}
		})
	}
}
