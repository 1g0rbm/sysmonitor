package metric

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdate(t *testing.T) {
	tests := []struct {
		name         string
		updDuration  int
		sendDuration int
		wantErr      error
	}{
		{
			name:         "update duration are greater then send duration test",
			updDuration:  10,
			sendDuration: 5,
			wantErr:      errors.New("update duration (10) should be less than send duration (5)"),
		},
		{
			name:         "update duration and send duration are equal test",
			updDuration:  5,
			sendDuration: 5,
			wantErr:      errors.New("update duration (5) should be less than send duration (5)"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Update(tt.updDuration, tt.sendDuration)

			assert.Equal(t, err, tt.wantErr)
		})
	}
}
