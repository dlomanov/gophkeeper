package usecases_test

import (
	"context"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	"github.com/dlomanov/gophkeeper/internal/entities"
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
	sut := usecases.NewEntryUC(
		zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel)),
		NewMockEntryRepo(),
		NewMockTrmManager())
	userID1 := uuid.New()
	userID2 := uuid.New()

	getAll, err := sut.GetEntries(ctx, usecases.GetEntriesRequest{UserID: userID1})
	require.NoError(t, err, "no error expected")
	require.Empty(t, getAll.Entries, "expected empty list")

	entries := make([]*entities.Entry, 3)
	entries[0], err = entities.NewEntry("key1", userID1, entities.EntryTypePassword, []byte("test_data_1"))
	require.NoError(t, err, "no error expected")
	entries[1], err = entities.NewEntry("key2", userID1, entities.EntryTypeBinary, []byte("test_data_2"))
	require.NoError(t, err, "no error expected")
	entries[2], err = entities.NewEntry("key3", userID1, entities.EntryTypeNote, []byte("test_data_3"))
	require.NoError(t, err, "no error expected")
	for i, entry := range entries {
		create, err := sut.Create(ctx, usecases.CreateEntryRequest{
			Key:    entry.Key,
			UserID: entry.UserID,
			Type:   entry.Type,
			Meta:   entry.Meta,
			Data:   entry.Data,
		})
		require.NoError(t, err, "no error expected")
		assert.NotEmpty(t, create.ID, "expected non-empty ID")
		assert.NotEmpty(t, create.CreatedAt, "expected non-empty created at")
		assert.NotEmpty(t, create.UpdatedAt, "expected non-empty updated at")
		assert.Equal(t, create.CreatedAt.Format("2006-01-02 15:04:05.000"), create.UpdatedAt.Format("2006-01-02 15:04:05.000"), "expected created at and updated at to be equal")
		entries[i].ID = create.ID
		entries[i].CreatedAt = create.CreatedAt
		entries[i].UpdatedAt = create.UpdatedAt
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
		assert.Equal(t, entries[i].CreatedAt.Format("2006-01-02 15:04:05.000"), entry.CreatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry created at")
		assert.Equal(t, entries[i].UpdatedAt.Format("2006-01-02 15:04:05.000"), entry.UpdatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry updated at")
	}

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
	assert.Equal(t, del.CreatedAt.Format("2006-01-02 15:04:05.000"), getAll.Entries[0].CreatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry created at")
	assert.Equal(t, del.UpdatedAt.Format("2006-01-02 15:04:05.000"), getAll.Entries[0].UpdatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry updated at")
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
		assert.Equal(t, entries[i].CreatedAt.Format("2006-01-02 15:04:05.000"), entry.CreatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry created at")
		assert.Equal(t, entries[i].UpdatedAt.Format("2006-01-02 15:04:05.000"), entry.UpdatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry updated at")
	}

	entries[0].Meta = map[string]string{"updated_test_key": "updated_test_value"}
	entries[0].Data = []byte("updated_test_data")
	update, err := sut.Update(ctx, usecases.UpdateEntryRequest{
		ID:     entries[0].ID,
		UserID: userID1,
		Meta:   entries[0].Meta,
		Data:   entries[0].Data,
	})
	require.NoError(t, err, "no error expected")
	assert.Equal(t, update.CreatedAt.Format("2006-01-02 15:04:05.000"), getAll.Entries[0].CreatedAt.Format("2006-01-02 15:04:05.000"), "expected updated at to be equal")
	assert.GreaterOrEqual(t, update.UpdatedAt, entries[0].UpdatedAt, "expected updated at to be changed")
	get, err := sut.Get(ctx, usecases.GetEntryRequest{ID: entries[0].ID, UserID: userID1})
	require.NoError(t, err, "no error expected")
	assert.Equal(t, get.Entry.ID.String(), update.ID.String(), "expected same entry")
	assert.Equal(t, get.Entry.Key, getAll.Entries[0].Key, "expected same entry keys")
	assert.Equal(t, get.Entry.UserID.String(), getAll.Entries[0].UserID.String(), "expected same user IDs")
	assert.Equal(t, get.Entry.Type, getAll.Entries[0].Type, "expected same entry types")
	assert.True(t, reflect.DeepEqual(get.Entry.Meta, entries[0].Meta), "expected same entry meta")
	assert.Equal(t, get.Entry.Data, entries[0].Data, "expected same entry data")
	assert.Equal(t, get.Entry.CreatedAt.Format("2006-01-02 15:04:05.000"), update.CreatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry created at")
	assert.Equal(t, get.Entry.UpdatedAt.Format("2006-01-02 15:04:05.000"), update.UpdatedAt.Format("2006-01-02 15:04:05.000"), "expected same entry updated at")
}

func TestEntryUC_validation(t *testing.T) {
	ctx := context.Background()
	sut := usecases.NewEntryUC(
		zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel)),
		NewMockEntryRepo(),
		NewMockTrmManager())

	largeData := []byte(strings.Repeat("s", entities.EntryMaxDataSize+1))

	_, err := sut.GetEntries(ctx, usecases.GetEntriesRequest{UserID: uuid.Nil})
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

	_, err = sut.Update(ctx, usecases.UpdateEntryRequest{ID: uuid.Nil, UserID: uuid.Nil, Meta: nil, Data: nil})
	require.ErrorIs(t, err, entities.ErrUserIDInvalid, "expected user ID invalid error")
	require.ErrorIs(t, err, entities.ErrEntryIDInvalid, "expected entry ID invalid error")
	require.ErrorIs(t, err, entities.ErrEntryDataEmpty, "expected entry data empty error")
	_, err = sut.Update(ctx, usecases.UpdateEntryRequest{ID: uuid.Nil, UserID: uuid.Nil, Meta: nil, Data: largeData})
	require.ErrorIs(t, err, entities.ErrUserIDInvalid, "expected user ID invalid error")
	require.ErrorIs(t, err, entities.ErrEntryIDInvalid, "expected entry ID invalid error")
	require.ErrorIs(t, err, entities.ErrEntryDataSizeExceeded, "expected entry data size exceeded error")

	_, err = sut.Delete(ctx, usecases.DeleteEntryRequest{ID: uuid.Nil, UserID: uuid.Nil})
	require.ErrorIs(t, err, entities.ErrUserIDInvalid, "expected user ID invalid error")
	require.ErrorIs(t, err, entities.ErrEntryIDInvalid, "expected entry ID invalid error")
}
