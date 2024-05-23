package entities_test

import (
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestNewEntry(t *testing.T) {
	tests := []struct {
		name    string
		typ     entities.EntryType
		data    []byte
		meta    map[string]string
		wantErr error
	}{
		{
			name:    "invalid type",
			typ:     "test",
			data:    nil,
			meta:    nil,
			wantErr: entities.ErrEntryTypeInvalid,
		},
		{
			name:    "invalid data",
			typ:     entities.EntryTypePassword,
			data:    nil,
			meta:    nil,
			wantErr: entities.ErrEntryDataInvalid,
		},
		{
			name: "valid without metadata",
			typ:  entities.EntryTypePassword,
			data: []byte("test"),
			meta: nil,
		},
		{
			name: "valid",
			typ:  entities.EntryTypePassword,
			data: []byte("test"),
			meta: map[string]string{"test": "test"},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			entry, err := entities.NewEntry(tt.typ, tt.data, tt.meta)
			if err != nil {
				require.ErrorIs(t, err, tt.wantErr, "error mismatch")
				require.Nil(t, entry, "entry should be nil on error")
				return
			}

			require.NotNil(t, entry, "entry should not be nil")
			assert.Equal(t, tt.typ, entry.Type, "type mismatch")
			if tt.data == nil {
				assert.Equal(t, tt.data, entry.Data, "data mismatch")
			}
			if tt.meta == nil {
				assert.Equal(t, tt.meta, entry.Metadata, "metadata mismatch")
			}
			assert.NotEmpty(t, entry.CreatedAt, "created at should not be empty")
			assert.NotEmpty(t, entry.UpdatedAt, "updated at should not be empty")
			assert.Equal(t, entry.CreatedAt, entry.UpdatedAt, "created at should be equal to updated at")
		})
	}
}
