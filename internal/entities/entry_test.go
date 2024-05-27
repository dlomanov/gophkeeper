package entities_test

import (
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewEntry(t *testing.T) {
	tests := []struct {
		userID  uuid.UUID
		name    string
		typ     entities.EntryType
		data    []byte
		meta    map[string]string
		wantErr error
	}{
		{
			name:    "invalid user ID",
			typ:     "test",
			data:    nil,
			meta:    nil,
			wantErr: entities.ErrUserIDInvalid,
		},
		{
			name:    "invalid type",
			userID:  uuid.New(),
			typ:     "test",
			data:    nil,
			meta:    nil,
			wantErr: entities.ErrEntryTypeInvalid,
		},
		{
			name:    "invalid data",
			userID:  uuid.New(),
			typ:     entities.EntryTypePassword,
			data:    nil,
			meta:    nil,
			wantErr: entities.ErrEntryDataEmpty,
		},
		{
			name:   "valid without metadata",
			userID: uuid.New(),
			typ:    entities.EntryTypePassword,
			data:   []byte("test"),
			meta:   map[string]string{"test": "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := entities.NewEntry(tt.userID, tt.typ, tt.data)
			if err != nil {
				require.ErrorIs(t, err, tt.wantErr, "error mismatch")
				require.Nil(t, entry, "entry should be nil on error")
				return
			}
			entry.Meta = tt.meta

			require.NotNil(t, entry, "entry should not be nil")
			assert.Equal(t, tt.typ, entry.Type, "type mismatch")
			if tt.data == nil {
				assert.Equal(t, tt.data, entry.Data, "data mismatch")
			}
			if tt.meta == nil {
				assert.Equal(t, tt.meta, entry.Meta, "metadata mismatch")
			}
			assert.NotEmpty(t, entry.CreatedAt, "created at should not be empty")
			assert.NotEmpty(t, entry.UpdatedAt, "updated at should not be empty")
			assert.Equal(t, entry.CreatedAt, entry.UpdatedAt, "created at should be equal to updated at")
		})
	}
}

func TestUpdate_nonOptions(t *testing.T) {
	entry, err := entities.NewEntry(uuid.New(), entities.EntryTypePassword, []byte("test"))
	require.NoError(t, err, "failed to create entry")
	err = entry.Update()
	require.NoError(t, err)
	assert.Equal(t, entry.CreatedAt, entry.UpdatedAt, "created at should be equal to updated at")
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name     string
		typ      entities.EntryType
		data     []byte
		meta     map[string]string
		wantErrs []error
	}{
		{
			name: "invalid type",
			typ:  "test",
			data: nil,
			meta: nil,
			wantErrs: []error{
				entities.ErrEntryTypeInvalid,
				entities.ErrEntryDataEmpty,
			},
		},
		{
			name: "invalid data",
			typ:  entities.EntryTypePassword,
			data: nil,
			meta: nil,
			wantErrs: []error{
				entities.ErrEntryDataEmpty,
			},
		},
		{
			name: "valid without metadata",
			typ:  entities.EntryTypePassword,
			data: []byte("test"),
			meta: nil,
		},
		{
			name: "valid with metadata",
			typ:  entities.EntryTypePassword,
			data: []byte("test"),
			meta: map[string]string{"test": "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now().UTC()
			entry := entities.Entry{
				UpdatedAt: now,
			}
			errs := entry.Update(
				entities.UpdateEntryType(tt.typ),
				entities.UpdateEntryData(tt.data),
				entities.UpdateEntryMeta(tt.meta))
			for _, wantErr := range tt.wantErrs {
				require.ErrorIs(t, errs, wantErr, "error mismatch")
				return
			}
			if len(tt.wantErrs) == 0 {
				require.NoError(t, errs, "error mismatch")
			}
			assert.Equal(t, tt.typ, entry.Type, "type mismatch")
			assert.Equal(t, tt.data, entry.Data, "data mismatch")
			assert.True(t, reflect.DeepEqual(tt.meta, entry.Meta), "metadata mismatch")
			assert.NotEmpty(t, entry.UpdatedAt, "updated at should not be empty")
			assert.GreaterOrEqual(t, entry.UpdatedAt, now, "updated at should be greater or equal to created at")
		})
	}
}

func TestUpdate_invalidOptions(t *testing.T) {
	entry, err := entities.NewEntry(uuid.New(), entities.EntryTypePassword, []byte("test"))
	require.NoError(t, err, "failed to create entry")
	errs := entry.Update(
		entities.UpdateEntryType("test"),
		entities.UpdateEntryData(nil),
		entities.UpdateEntryData([]byte(strings.Repeat("s", entities.EntryMaxDataSize+1))),
	)
	require.ErrorIs(t, errs, entities.ErrEntryTypeInvalid, "want type invalid error")
	require.ErrorIs(t, errs, entities.ErrEntryDataEmpty, "want data empty error")
	require.ErrorIs(t, errs, entities.ErrEntryDataSizeExceeded, "want data size exceeded error")
}
