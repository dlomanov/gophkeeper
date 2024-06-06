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
		testName string
		key      string
		userID   uuid.UUID
		typ      entities.EntryType
		data     []byte
		meta     map[string]string
		wantErrs []error
	}{
		{
			testName: "all invalid",
			key:      "",
			userID:   uuid.Nil,
			typ:      "test",
			data:     nil,
			meta:     nil,
			wantErrs: []error{
				entities.ErrEntryKeyInvalid,
				entities.ErrUserIDInvalid,
				entities.ErrEntryTypeInvalid,
				entities.ErrEntryDataEmpty,
			},
		},
		{
			testName: "all invalid but key",
			key:      "key",
			userID:   uuid.Nil,
			typ:      "test",
			data:     nil,
			meta:     nil,
			wantErrs: []error{
				entities.ErrUserIDInvalid,
				entities.ErrEntryTypeInvalid,
				entities.ErrEntryDataEmpty,
			},
		},
		{
			testName: "all invalid but key, userID",
			key:      "key",
			userID:   uuid.New(),
			typ:      "test",
			data:     nil,
			meta:     nil,
			wantErrs: []error{
				entities.ErrEntryTypeInvalid,
				entities.ErrEntryDataEmpty,
			},
		},
		{
			testName: "all invalid but key, userID, typ",
			key:      "key",
			userID:   uuid.New(),
			typ:      entities.EntryTypeBinary,
			data:     nil,
			meta:     nil,
			wantErrs: []error{
				entities.ErrEntryDataEmpty,
			},
		},
		{
			testName: "almost valid but data size exceeded",
			key:      "key",
			userID:   uuid.New(),
			typ:      entities.EntryTypeBinary,
			data:     []byte(strings.Repeat("s", entities.EntryMaxDataSize+1)),
			meta:     nil,
			wantErrs: []error{entities.ErrEntryDataSizeExceeded},
		},
		{
			testName: "valid without metadata",
			key:      "key",
			userID:   uuid.New(),
			typ:      entities.EntryTypeBinary,
			data:     []byte("data"),
			meta:     nil,
			wantErrs: nil,
		},
		{
			testName: "valid without metadata",
			key:      "key",
			userID:   uuid.New(),
			typ:      entities.EntryTypeBinary,
			data:     []byte("data"),
			meta:     map[string]string{"key": "value"},
			wantErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			entry, err := entities.NewEntry(tt.key, tt.userID, tt.typ, tt.data)
			if len(tt.wantErrs) != 0 {
				for _, wantErr := range tt.wantErrs {
					assert.ErrorIs(t, err, wantErr, "unexpected error")
				}
				return
			}
			require.NoError(t, err, "no error expected")
			entry.Meta = tt.meta

			require.NotNil(t, entry, "entry should not be nil")
			assert.Equal(t, tt.key, entry.Key, "key mismatch")
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

func TestUpdateVersion(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		meta     map[string]string
		wantErrs []error
	}{
		{
			name: "invalid data",
			data: nil,
			meta: nil,
			wantErrs: []error{
				entities.ErrEntryDataEmpty,
			},
		},
		{
			name: "valid without metadata",
			data: []byte("test"),
			meta: nil,
		},
		{
			name: "valid with metadata",
			data: []byte("test"),
			meta: map[string]string{"test": "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now().UTC()
			typ := entities.EntryTypeBinary
			entry := entities.Entry{
				Version:   1,
				Type:      typ,
				UpdatedAt: now,
			}
			srcVersion := entry.Version
			errs := entry.UpdateVersion(
				entry.Version,
				entities.UpdateEntryData(tt.data),
				entities.UpdateEntryMeta(tt.meta))
			for _, wantErr := range tt.wantErrs {
				require.ErrorIs(t, errs, wantErr, "error mismatch")
				return
			}
			if len(tt.wantErrs) == 0 {
				require.NoError(t, errs, "error mismatch")
			}
			assert.Equal(t, typ, entry.Type, "type mismatch")
			assert.Equal(t, tt.data, entry.Data, "data mismatch")
			assert.True(t, reflect.DeepEqual(tt.meta, entry.Meta), "metadata mismatch")
			assert.NotEmpty(t, entry.UpdatedAt, "updated at should not be empty")
			assert.GreaterOrEqual(t, entry.UpdatedAt, now, "updated at should be greater or equal to created at")
			assert.Equal(t, srcVersion+1, entry.Version, "version mismatch")
		})
	}
}

func TestUpdateVersion_nonOptions(t *testing.T) {
	entry, err := entities.NewEntry("key", uuid.New(), entities.EntryTypePassword, []byte("test"))
	require.NoError(t, err, "failed to create entry")
	err = entry.UpdateVersion(entry.Version)
	require.NoError(t, err)
	assert.Equal(t, entry.CreatedAt, entry.UpdatedAt, "created at should be equal to updated at")
}

func TestUpdateVersion_invalidOptions(t *testing.T) {
	entry, err := entities.NewEntry("key", uuid.New(), entities.EntryTypePassword, []byte("test"))
	require.NoError(t, err, "failed to create entry")
	errs := entry.UpdateVersion(entry.Version,
		entities.UpdateEntryData(nil),
		entities.UpdateEntryData([]byte(strings.Repeat("s", entities.EntryMaxDataSize+1))),
	)
	require.ErrorIs(t, errs, entities.ErrEntryDataEmpty, "want data empty error")
	require.ErrorIs(t, errs, entities.ErrEntryDataSizeExceeded, "want data size exceeded error")
}

func TestUpdateVersion_versionConflict(t *testing.T) {
	entry, err := entities.NewEntry("key", uuid.New(), entities.EntryTypePassword, []byte("test"))
	require.NoError(t, err, "failed to create entry")
	errs := entry.UpdateVersion(entry.Version-1, entities.UpdateEntryData([]byte("test1")))
	require.ErrorIs(t, errs, entities.ErrEntryVersionConflict, "want version conflict error")
}
