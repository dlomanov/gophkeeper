package usecases_test

import (
	"context"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/dlomanov/gophkeeper/internal/infra/encrypto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"reflect"
	"strings"
	"testing"
)

func TestEntryUC(t *testing.T) {
	ctx := context.Background()
	enc, err := encrypto.NewEncrypter([]byte("1234567890123456"))
	require.NoError(t, err, "no error expected")
	sut := usecases.NewEntryUC(
		zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel)),
		NewMockEntryRepo(),
		enc,
		NewMockTrmManager())
	userID1 := uuid.New()
	userID2 := uuid.New()

	// GetEntries (empty)
	getAll, err := sut.GetEntries(ctx, usecases.GetEntriesRequest{UserID: userID1})
	require.NoError(t, err, "no error expected")
	require.Empty(t, getAll.Entries, "expected empty list")

	// Create + GetEntries
	entries := make([]*entities.Entry, 3)
	entries[0], err = entities.NewEntry("key1", userID1, entities.EntryTypePassword, []byte("test_data_1"))
	require.NoError(t, err, "no error expected")
	entries[1], err = entities.NewEntry("key2", userID1, entities.EntryTypeBinary, []byte("test_data_2"))
	require.NoError(t, err, "no error expected")
	entries[2], err = entities.NewEntry("key3", userID1, entities.EntryTypeNote, []byte("test_data_3"))
	require.NoError(t, err, "no error expected")
	for i, entry := range entries {
		created, err := sut.Create(ctx, usecases.CreateEntryRequest{
			Key:    entry.Key,
			UserID: entry.UserID,
			Type:   entry.Type,
			Meta:   entry.Meta,
			Data:   entry.Data,
		})
		require.NoError(t, err, "no error expected")
		assert.NotEmpty(t, created.ID, "expected non-empty ID")
		assert.Equal(t, created.Version, int64(1), "expected version 1 after creation")
		entries[i].ID = created.ID
		entries[i].Version = created.Version
	}
	_, err = sut.Create(ctx, usecases.CreateEntryRequest{
		Key:    entries[0].Key,
		UserID: userID1,
		Type:   entities.EntryTypeNote,
		Meta:   map[string]string{"description": "test_note_4"},
		Data:   []byte("test_data_4"),
	})
	require.ErrorIs(t, err, entities.ErrEntryExists, "expected entry exists error")
	getAll, err = sut.GetEntries(ctx, usecases.GetEntriesRequest{UserID: userID1})
	require.NoError(t, err, "no error expected")
	require.NotEmpty(t, getAll.Entries, "expected non-empty list")
	for i, entry := range getAll.Entries {
		assert.Equal(t, entries[i].Key, entry.Key, "expected same entry keys")
		assert.Equal(t, entries[i].UserID, entry.UserID, "expected same user IDs")
		assert.Equal(t, entries[i].Type, entry.Type, "expected same entry types")
		assert.True(t, reflect.DeepEqual(entries[i].Meta, entry.Meta), "expected same entry meta")
		assert.Equal(t, entries[i].Data, entry.Data, "expected same entry data")
		assert.Equal(t, entries[i].Version, entry.Version, "expected same entry versions")
		assert.Equal(t, entries[i].CreatedAt.Format("2006-01-02 15:04:05.000"), entry.CreatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry created at")
		assert.Equal(t, entries[i].UpdatedAt.Format("2006-01-02 15:04:05.000"), entry.UpdatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry updated at")
	}

	// Delete + GetEntries
	_, err = sut.Delete(ctx, usecases.DeleteEntryRequest{
		ID:     uuid.New(),
		UserID: userID2,
	})
	require.ErrorIs(t, err, entities.ErrEntryNotFound, "expected entry not found error")
	_, err = sut.Delete(ctx, usecases.DeleteEntryRequest{
		ID:     entries[0].ID,
		UserID: userID2,
	})
	require.ErrorIs(t, err, entities.ErrEntryNotFound, "expected entry not found error")
	del, err := sut.Delete(ctx, usecases.DeleteEntryRequest{
		ID:     entries[0].ID,
		UserID: userID1,
	})
	require.NoError(t, err, "no error expected")
	assert.Equal(t, del.ID.String(), getAll.Entries[0].ID.String(), "expected same entry IDs")
	assert.Equal(t, del.Version, getAll.Entries[0].Version, "expected same entry versions")
	entries = entries[1:]
	getAll, err = sut.GetEntries(ctx, usecases.GetEntriesRequest{UserID: userID1})
	require.NoError(t, err, "no error expected")
	require.NotEmpty(t, getAll.Entries, "expected non-empty list")
	for i, entry := range getAll.Entries {
		assert.Equal(t, entries[i].Key, entry.Key, "expected same entry keys")
		assert.Equal(t, entries[i].UserID, entry.UserID, "expected same user IDs")
		assert.Equal(t, entries[i].Type, entry.Type, "expected same entry types")
		assert.True(t, reflect.DeepEqual(entries[i].Meta, entry.Meta), "expected same entry meta")
		assert.Equal(t, entries[i].Data, entry.Data, "expected same entry data")
		assert.Equal(t, entries[i].Version, entry.Version, "expected same entry versions")
		assert.Equal(t, entries[i].CreatedAt.Format("2006-01-02 15:04:05.000"), entry.CreatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry created at")
		assert.Equal(t, entries[i].UpdatedAt.Format("2006-01-02 15:04:05.000"), entry.UpdatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry updated at")
	}

	// Update + Get
	entries[0].Meta = map[string]string{"updated_test_key": "updated_test_value"}
	entries[0].Data = []byte("updated_test_data")
	updated, err := sut.Update(ctx, usecases.UpdateEntryRequest{
		ID:      entries[0].ID,
		UserID:  userID1,
		Version: entries[0].Version,
		Meta:    entries[0].Meta,
		Data:    entries[0].Data,
	})
	require.NoError(t, err, "no error expected")
	assert.Equal(t, entries[0].Version+1, updated.Version, "expected updated version")
	entries[0].Version = updated.Version
	get, err := sut.Get(ctx, usecases.GetEntryRequest{ID: entries[0].ID, UserID: userID1})
	require.NoError(t, err, "no error expected")
	assert.Equal(t, get.Entry.ID.String(), updated.ID.String(), "expected same entry")
	assert.Equal(t, get.Entry.Key, getAll.Entries[0].Key, "expected same entry keys")
	assert.Equal(t, get.Entry.UserID.String(), getAll.Entries[0].UserID.String(), "expected same user IDs")
	assert.Equal(t, get.Entry.Type, getAll.Entries[0].Type, "expected same entry types")
	assert.True(t, reflect.DeepEqual(get.Entry.Meta, entries[0].Meta), "expected same entry meta")
	assert.Equal(t, get.Entry.Data, entries[0].Data, "expected same entry data")
	assert.Equal(t, get.Entry.Version, updated.Version, "expected same entry version")

	// GetNewestEntries
	getAll, err = sut.GetEntries(ctx, usecases.GetEntriesRequest{UserID: userID1})
	require.NoError(t, err, "no error expected")
	require.NotEmpty(t, getAll.Entries, "expected non-empty list")
	entries[0].Meta = map[string]string{"updated_test_key": "updated_test_value"}
	entries[0].Data = []byte("updated_test_data")
	updated, err = sut.Update(ctx, usecases.UpdateEntryRequest{
		ID:      entries[0].ID,
		UserID:  userID1,
		Version: entries[0].Version,
		Meta:    entries[0].Meta,
		Data:    entries[0].Data,
	})
	require.NoError(t, err, "no error expected")
	entries[0].Version = updated.Version
	versions := make(map[string]int64, len(getAll.Entries))
	for _, v := range getAll.Entries {
		versions[v.ID.String()] = v.Version
	}
	newest, err := sut.GetNewestEntries(ctx, usecases.GetNewestEntriesRequest{UserID: userID1, Versions: versions})
	require.NoError(t, err, "no error expected")
	require.Len(t, newest.Entries, 1, "expected one entry")
	assert.Equal(t, newest.Entries[0].ID.String(), entries[0].ID.String(), "expected same entry")
	assert.Equal(t, newest.Entries[0].Key, entries[0].Key, "expected same entry keys")
	assert.Equal(t, newest.Entries[0].UserID.String(), entries[0].UserID.String(), "expected same user IDs")
	assert.Equal(t, newest.Entries[0].Type, entries[0].Type, "expected same entry types")
	assert.True(t, reflect.DeepEqual(newest.Entries[0].Meta, entries[0].Meta), "expected same entry meta")
	assert.Equal(t, newest.Entries[0].Data, entries[0].Data, "expected same entry data")
	assert.Equal(t, newest.Entries[0].Version, entries[0].Version, "expected same entry version")

	// GetNewestEntries (empty versions)
	versions = nil
	newest, err = sut.GetNewestEntries(ctx, usecases.GetNewestEntriesRequest{UserID: userID1, Versions: versions})
	require.NoError(t, err, "no error expected")
	require.Equal(t, len(getAll.Entries), len(newest.Entries), "expected one entry")
}

func TestEntryUC_validation(t *testing.T) {
	ctx := context.Background()
	enc, err := encrypto.NewEncrypter([]byte("1234567890123456"))
	require.NoError(t, err, "no error expected")
	sut := usecases.NewEntryUC(
		zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel)),
		NewMockEntryRepo(),
		enc,
		NewMockTrmManager())

	largeData := []byte(strings.Repeat("s", entities.EntryMaxDataSize+1))

	_, err = sut.GetEntries(ctx, usecases.GetEntriesRequest{UserID: uuid.Nil})
	require.ErrorIs(t, err, entities.ErrUserIDInvalid, "expected user ID invalid error")

	_, err = sut.GetNewestEntries(ctx, usecases.GetNewestEntriesRequest{UserID: uuid.Nil, Versions: nil})
	require.ErrorIs(t, err, entities.ErrUserIDInvalid, "expected user ID invalid error")

	_, err = sut.Get(ctx, usecases.GetEntryRequest{ID: uuid.Nil, UserID: uuid.Nil})
	require.ErrorIs(t, err, entities.ErrUserIDInvalid, "expected user ID invalid error")
	require.ErrorIs(t, err, entities.ErrEntryIDInvalid, "expected entry ID invalid error")

	_, err = sut.Create(ctx, usecases.CreateEntryRequest{
		Key:    "",
		Type:   "",
		UserID: uuid.Nil,
		Meta:   nil,
		Data:   nil,
	})
	require.ErrorIs(t, err, entities.ErrEntryKeyInvalid, "expected entry key invalid error")
	require.ErrorIs(t, err, entities.ErrEntryTypeInvalid, "expected entry type invalid error")
	require.ErrorIs(t, err, entities.ErrUserIDInvalid, "expected user ID invalid error")
	require.ErrorIs(t, err, entities.ErrEntryDataEmpty, "expected entry data empty error")
	_, err = sut.Create(ctx, usecases.CreateEntryRequest{
		Key:    "",
		Type:   "",
		UserID: uuid.Nil,
		Meta:   nil,
		Data:   largeData,
	})
	require.ErrorIs(t, err, entities.ErrEntryKeyInvalid, "expected entry key invalid error")
	require.ErrorIs(t, err, entities.ErrEntryTypeInvalid, "expected entry type invalid error")
	require.ErrorIs(t, err, entities.ErrUserIDInvalid, "expected user ID invalid error")
	require.ErrorIs(t, err, entities.ErrEntryDataSizeExceeded, "expected entry data size exceeded error")

	_, err = sut.Update(ctx, usecases.UpdateEntryRequest{ID: uuid.Nil, UserID: uuid.Nil, Meta: nil, Data: nil, Version: 0})
	require.ErrorIs(t, err, entities.ErrUserIDInvalid, "expected user ID invalid error")
	require.ErrorIs(t, err, entities.ErrEntryIDInvalid, "expected entry ID invalid error")
	require.ErrorIs(t, err, entities.ErrEntryDataEmpty, "expected entry data empty error")
	require.ErrorIs(t, err, entities.ErrEntryVersionInvalid, "expected entry version invalid error")
	_, err = sut.Update(ctx, usecases.UpdateEntryRequest{ID: uuid.Nil, UserID: uuid.Nil, Meta: nil, Data: largeData, Version: 0})
	require.ErrorIs(t, err, entities.ErrUserIDInvalid, "expected user ID invalid error")
	require.ErrorIs(t, err, entities.ErrEntryIDInvalid, "expected entry ID invalid error")
	require.ErrorIs(t, err, entities.ErrEntryDataSizeExceeded, "expected entry data size exceeded error")
	require.ErrorIs(t, err, entities.ErrEntryVersionInvalid, "expected entry version invalid error")

	_, err = sut.Delete(ctx, usecases.DeleteEntryRequest{ID: uuid.Nil, UserID: uuid.Nil})
	require.ErrorIs(t, err, entities.ErrUserIDInvalid, "expected user ID invalid error")
	require.ErrorIs(t, err, entities.ErrEntryIDInvalid, "expected entry ID invalid error")
}
