package usecases_test

import (
	"context"
	"github.com/dlomanov/gophkeeper/internal/apps/server/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/services/diff"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/dlomanov/gophkeeper/internal/infra/encrypto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"strings"
	"testing"
)

func TestEntryUC_GetAll(t *testing.T) {
	var (
		ctx     = context.Background()
		sut     = createSUT(t)
		userID1 = uuid.New()
		userID2 = uuid.New()
	)

	empty := func(t require.TestingT, request any, response any, args ...any) {
		resp := response.(entities.GetEntriesResponse)
		require.Empty(t, resp.Entries, args...)
	}
	exactErr := func(target error) require.ErrorAssertionFunc {
		return func(t require.TestingT, err error, args ...any) {
			require.ErrorIs(t, err, target, args...)
		}
	}
	requests := []entities.CreateEntryRequest{
		{
			Key:    "key1",
			UserID: userID1,
			Type:   core.EntryTypeNote,
			Meta:   map[string]string{"description": "test_note_1"},
			Data:   []byte("test_data_1"),
		},
		{
			Key:    "key2",
			UserID: userID1,
			Type:   core.EntryTypeNote,
			Meta:   map[string]string{"description": "test_note_2"},
			Data:   []byte("test_data_2"),
		},
	}
	responses := make(map[uuid.UUID]entities.CreateEntryRequest, len(requests))
	for _, req := range requests {
		resp, err := sut.Create(ctx, req)
		require.NoError(t, err)
		responses[resp.ID] = req
	}

	tests := []struct {
		name         string
		request      entities.GetEntriesRequest
		wantErr      require.ErrorAssertionFunc
		wantResponse require.ComparisonAssertionFunc
	}{
		{
			name: "invalid user ID",
			request: entities.GetEntriesRequest{
				UserID: uuid.Nil,
			},
			wantErr:      exactErr(entities.ErrUserIDInvalid),
			wantResponse: empty,
		},
		{
			name: "empty",
			request: entities.GetEntriesRequest{
				UserID: userID2,
			},
			wantErr:      require.NoError,
			wantResponse: empty,
		},
		{
			name: "all",
			request: entities.GetEntriesRequest{
				UserID: userID1,
			},
			wantErr: require.NoError,
			wantResponse: func(t require.TestingT, request any, response any, _ ...any) {
				req := request.(entities.GetEntriesRequest)
				resp := response.(entities.GetEntriesResponse)
				require.Len(t, resp.Entries, len(responses))
				for _, entry := range resp.Entries {
					created, ok := responses[entry.ID]
					require.True(t, ok, "entry should exist")
					require.Equal(t, created.Key, entry.Key)
					require.Equal(t, created.Type, entry.Type)
					require.Equal(t, created.Data, entry.Data)
					require.Equal(t, created.Meta, entry.Meta)
					require.Equal(t, created.UserID, entry.UserID)
					require.Equal(t, req.UserID, entry.UserID)
					require.Equal(t, int64(1), entry.Version)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := sut.GetEntries(ctx, tt.request)
			tt.wantErr(t, err)
			tt.wantResponse(t, tt.request, resp)
		})
	}
}

func TestEntryUC_GetEntriesDiff(t *testing.T) {
	var (
		ctx     = context.Background()
		sut     = createSUT(t)
		userID1 = uuid.New()
		userID2 = uuid.New()
	)

	empty := func(t require.TestingT, request any, response any, args ...any) {
		resp := response.(entities.GetEntriesDiffResponse)
		require.Empty(t, resp.Entries, args...)
		require.Empty(t, resp.CreateIDs, args...)
		require.Empty(t, resp.UpdateIDs, args...)
		require.Empty(t, resp.DeleteIDs, args...)
	}
	exactErr := func(target error) require.ErrorAssertionFunc {
		return func(t require.TestingT, err error, args ...any) {
			require.ErrorIs(t, err, target, args...)
		}
	}
	requests := []entities.CreateEntryRequest{
		{
			Key:    "key1",
			UserID: userID1,
			Type:   core.EntryTypeNote,
			Meta:   map[string]string{"description": "test_note_1"},
			Data:   []byte("test_data_1"),
		},
		{
			Key:    "key2",
			UserID: userID1,
			Type:   core.EntryTypeNote,
			Meta:   map[string]string{"description": "test_note_2"},
			Data:   []byte("test_data_2"),
		},
		{
			Key:    "key3",
			UserID: userID1,
			Type:   core.EntryTypeNote,
			Meta:   map[string]string{"description": "test_note_3"},
			Data:   []byte("test_data_3"),
		},
	}
	responses := make(map[uuid.UUID]entities.CreateEntryRequest, len(requests))
	for _, req := range requests {
		resp, err := sut.Create(ctx, req)
		require.NoError(t, err)
		responses[resp.ID] = req
	}
	versions := make([]core.EntryVersion, 0, len(responses))
	for k := range responses {
		versions = append(versions, core.EntryVersion{ID: k, Version: 1})
	}
	clone := func(
		versions []core.EntryVersion,
		fn func(v []core.EntryVersion) []core.EntryVersion,
	) []core.EntryVersion {
		v := make([]core.EntryVersion, len(versions))
		copy(v, versions)
		return fn(v)
	}

	tests := []struct {
		name         string
		request      entities.GetEntriesDiffRequest
		pretest      func()
		wantErr      require.ErrorAssertionFunc
		wantResponse require.ComparisonAssertionFunc
	}{
		{
			name: "invalid user ID",
			request: entities.GetEntriesDiffRequest{
				UserID:   uuid.Nil,
				Versions: versions,
			},
			wantErr:      exactErr(entities.ErrUserIDInvalid),
			wantResponse: empty,
		},
		{
			name: "delete all",
			request: entities.GetEntriesDiffRequest{
				UserID:   userID2,
				Versions: versions,
			},
			wantErr: require.NoError,
			wantResponse: func(t require.TestingT, request any, response any, _ ...any) {
				req := request.(entities.GetEntriesDiffRequest)
				resp := response.(entities.GetEntriesDiffResponse)
				require.Empty(t, resp.Entries)
				require.Empty(t, resp.CreateIDs)
				require.Empty(t, resp.UpdateIDs)
				require.Len(t, resp.DeleteIDs, len(req.Versions))
				del := make(map[uuid.UUID]struct{}, len(resp.DeleteIDs))
				for _, id := range resp.DeleteIDs {
					del[id] = struct{}{}
				}
				for _, version := range req.Versions {
					_, ok := del[version.ID]
					require.True(t, ok, "version should match with deleted")
				}
			},
		},
		{
			name: "diff",
			request: entities.GetEntriesDiffRequest{
				UserID: userID1,
				Versions: clone(versions, func(v []core.EntryVersion) []core.EntryVersion {
					v[0].Version = 2
					v[len(v)-1].ID = uuid.New()
					return v
				}),
			},
			wantErr: require.NoError,
			wantResponse: func(t require.TestingT, request any, response any, _ ...any) {
				req := request.(entities.GetEntriesDiffRequest)
				resp := response.(entities.GetEntriesDiffResponse)
				require.Len(t, resp.Entries, 2)
				require.Len(t, resp.CreateIDs, 1)
				require.Len(t, resp.UpdateIDs, 1)
				require.Len(t, resp.DeleteIDs, 1)
				require.Equal(t, resp.DeleteIDs[0], req.Versions[len(req.Versions)-1].ID)
				require.Equal(t, resp.UpdateIDs[0], req.Versions[0].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pretest != nil {
				tt.pretest()
			}
			resp, err := sut.GetEntriesDiff(ctx, tt.request)
			tt.wantErr(t, err)
			tt.wantResponse(t, tt.request, resp)
		})
	}
}

func TestEntryUC_Create(t *testing.T) {
	var (
		ctx     = context.Background()
		sut     = createSUT(t)
		userID1 = uuid.New()
		userID2 = uuid.New()
	)

	okResponse := func(t require.TestingT, request any, response any, args ...any) {
		req := request.(entities.CreateEntryRequest)
		resp := response.(entities.CreateEntryResponse)
		getResp, err := sut.Get(ctx, entities.GetEntryRequest{
			ID:     resp.ID,
			UserID: req.UserID,
		})
		require.NoError(t, err)
		require.Equal(t, req.Key, getResp.Entry.Key)
		require.Equal(t, req.UserID, getResp.Entry.UserID)
		require.Equal(t, req.Type, getResp.Entry.Type)
		require.Equal(t, req.Meta, getResp.Entry.Meta)
		require.Equal(t, req.Data, getResp.Entry.Data)
		require.Equal(t, resp.ID, getResp.Entry.ID)
		require.Equal(t, resp.Version, getResp.Entry.Version)
	}

	tests := []struct {
		name         string
		request      entities.CreateEntryRequest
		wantErr      require.ErrorAssertionFunc
		wantResponse require.ComparisonAssertionFunc
	}{
		{
			name: "key1 user1",
			request: entities.CreateEntryRequest{
				Key:    "key1",
				UserID: userID1,
				Type:   core.EntryTypePassword,
				Meta:   nil,
				Data:   []byte("test_data_1"),
			},
			wantErr:      require.NoError,
			wantResponse: okResponse,
		},
		{
			name: "key1 user2",
			request: entities.CreateEntryRequest{
				Key:    "key1",
				UserID: userID2,
				Type:   core.EntryTypePassword,
				Meta:   nil,
				Data:   []byte("test_data_1"),
			},
			wantErr:      require.NoError,
			wantResponse: okResponse,
		},
		{
			name: "key1 user2 conflict",
			request: entities.CreateEntryRequest{
				Key:    "key1",
				UserID: userID2,
				Type:   core.EntryTypePassword,
				Meta:   nil,
				Data:   []byte("test_data_1"),
			},
			wantErr: require.NoError,
			wantResponse: func(t require.TestingT, request any, response any, args ...any) {
				req := request.(entities.CreateEntryRequest)
				resp := response.(entities.CreateEntryResponse)
				getResp, err := sut.Get(ctx, entities.GetEntryRequest{
					ID:     resp.ID,
					UserID: req.UserID,
				})
				require.NoError(t, err)
				require.NotEqual(t, req.Key, getResp.Entry.Key)
				require.True(t, strings.HasPrefix(getResp.Entry.Key, req.Key), "key should be prefix of %s", req.Key)
				require.Equal(t, req.UserID, getResp.Entry.UserID)
				require.Equal(t, req.Type, getResp.Entry.Type)
				require.Equal(t, req.Meta, getResp.Entry.Meta)
				require.Equal(t, req.Data, getResp.Entry.Data)
				require.Equal(t, resp.ID, getResp.Entry.ID)
				require.Equal(t, resp.Version, getResp.Entry.Version)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := sut.Create(ctx, tt.request)
			tt.wantErr(t, err)
			tt.wantResponse(t, tt.request, response)
		})
	}
}

func TestEntryUC_Delete(t *testing.T) {
	var (
		ctx     = context.Background()
		sut     = createSUT(t)
		userID1 = uuid.New()
		userID2 = uuid.New()
	)

	nothing := func(t require.TestingT, request any, response any, args ...any) {}
	exactErr := func(target error) require.ErrorAssertionFunc {
		return func(t require.TestingT, err error, args ...any) {
			require.ErrorIs(t, err, target, args...)
		}
	}

	createResponse, err := sut.Create(ctx, entities.CreateEntryRequest{
		Key:    "key1",
		UserID: userID1,
		Type:   core.EntryTypeNote,
		Meta:   map[string]string{"description": "test_note_1"},
		Data:   []byte("test_data_1"),
	})
	require.NoError(t, err)

	tests := []struct {
		name         string
		request      entities.DeleteEntryRequest
		wantErr      require.ErrorAssertionFunc
		wantResponse require.ComparisonAssertionFunc
	}{
		{
			name: "not found with random ID",
			request: entities.DeleteEntryRequest{
				ID:     uuid.New(),
				UserID: userID1,
			},
			wantErr:      exactErr(entities.ErrEntryNotFound),
			wantResponse: nothing,
		},
		{
			name: "not found with wrong user ID",
			request: entities.DeleteEntryRequest{
				ID:     createResponse.ID,
				UserID: userID2,
			},
			wantErr:      exactErr(entities.ErrEntryNotFound),
			wantResponse: nothing,
		},
		{
			name: "deleted",
			request: entities.DeleteEntryRequest{
				ID:     createResponse.ID,
				UserID: userID1,
			},
			wantErr: require.NoError,
			wantResponse: func(t require.TestingT, request any, response any, args ...any) {
				req := request.(entities.DeleteEntryRequest)
				resp := response.(entities.DeleteEntryResponse)
				require.Equal(t, req.ID, resp.ID)
				require.Equal(t, createResponse.ID, resp.ID)
				require.Equal(t, createResponse.Version, resp.Version)
				_, err := sut.Get(ctx, entities.GetEntryRequest{ID: req.ID, UserID: req.UserID})
				require.ErrorIs(t, err, entities.ErrEntryNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := sut.Delete(ctx, tt.request)
			tt.wantErr(t, err)
			tt.wantResponse(t, tt.request, resp)
		})
	}
}

func TestEntryUC_Update(t *testing.T) {
	var (
		ctx     = context.Background()
		sut     = createSUT(t)
		userID1 = uuid.New()
		userID2 = uuid.New()
	)

	nothing := func(t require.TestingT, request any, response any, args ...any) {}
	exactErr := func(target error) require.ErrorAssertionFunc {
		return func(t require.TestingT, err error, args ...any) {
			require.ErrorIs(t, err, target, args...)
		}
	}

	createResponse, err := sut.Create(ctx, entities.CreateEntryRequest{
		Key:    "key1",
		UserID: userID1,
		Type:   core.EntryTypeNote,
		Meta:   map[string]string{"description": "test_note_1"},
		Data:   []byte("test_data_1"),
	})
	require.NoError(t, err)
	updatedMeta := map[string]string{"description": "test_note_2"}
	updatedData := []byte("test_data_2")

	tests := []struct {
		name         string
		request      entities.UpdateEntryRequest
		wantErr      require.ErrorAssertionFunc
		wantResponse require.ComparisonAssertionFunc
	}{
		{
			name: "not found with random ID",
			request: entities.UpdateEntryRequest{
				ID:      uuid.New(),
				Meta:    updatedMeta,
				Data:    updatedData,
				Version: createResponse.Version,
				UserID:  userID1,
			},
			wantErr:      exactErr(entities.ErrEntryNotFound),
			wantResponse: nothing,
		},
		{
			name: "not found with wrong user ID",
			request: entities.UpdateEntryRequest{
				ID:      createResponse.ID,
				Meta:    updatedMeta,
				Data:    updatedData,
				Version: createResponse.Version,
				UserID:  userID2,
			},
			wantErr:      exactErr(entities.ErrEntryNotFound),
			wantResponse: nothing,
		},
		{
			name: "update",
			request: entities.UpdateEntryRequest{
				ID:      createResponse.ID,
				Meta:    updatedMeta,
				Data:    updatedData,
				Version: createResponse.Version,
				UserID:  userID1,
			},
			wantErr: require.NoError,
			wantResponse: func(t require.TestingT, request any, response any, args ...any) {
				req := request.(entities.UpdateEntryRequest)
				resp := response.(entities.UpdateEntryResponse)
				require.Equal(t, req.ID, resp.ID)
				require.Equal(t, createResponse.ID, resp.ID)
				require.Equal(t, createResponse.Version+1, resp.Version)
				getResp, err := sut.Get(ctx, entities.GetEntryRequest{ID: resp.ID, UserID: req.UserID})
				require.NoError(t, err)
				require.Equal(t, req.Meta, getResp.Entry.Meta)
				require.Equal(t, req.Data, getResp.Entry.Data)
			},
		},
		{
			name: "update conflict",
			request: entities.UpdateEntryRequest{
				ID:      createResponse.ID,
				Meta:    map[string]string{"description": "conflict"},
				Data:    []byte("conflict"),
				Version: createResponse.Version + 2,
				UserID:  userID1,
			},
			wantErr: require.NoError,
			wantResponse: func(t require.TestingT, request any, response any, args ...any) {
				req := request.(entities.UpdateEntryRequest)
				resp := response.(entities.UpdateEntryResponse)
				require.NotEqual(t, req.ID, resp.ID)
				require.Equal(t, int64(1), resp.Version)
				getResp, err := sut.Get(ctx, entities.GetEntryRequest{ID: resp.ID, UserID: req.UserID})
				require.NoError(t, err)
				require.Equal(t, req.Meta, getResp.Entry.Meta)
				require.Equal(t, req.Data, getResp.Entry.Data)
				require.NotEqual(t, req.Version, getResp.Entry.Version)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := sut.Update(ctx, tt.request)
			tt.wantErr(t, err)
			tt.wantResponse(t, tt.request, resp)
		})
	}
}

func createSUT(t *testing.T) *usecases.EntryUC {
	merger := diff.NewEntry()
	enc, err := encrypto.NewEncrypter([]byte("1234567890123456"))
	require.NoError(t, err, "no error expected")
	return usecases.NewEntryUC(
		zaptest.NewLogger(t, zaptest.Level(zap.FatalLevel)),
		NewMockEntryRepo(),
		merger,
		enc,
		NewMockTrmManager())
}
