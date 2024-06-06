package diff

import (
	"context"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	uuid1 = uuid.New()
	uuid2 = uuid.New()
	uuid3 = uuid.New()
	uuid4 = uuid.New()
)

func TestEntry_GetDiff(t *testing.T) {
	tests := []struct {
		name          string
		server        []entities.EntryVersion
		client        []entities.EntryVersion
		wantCreateIDs []uuid.UUID
		wantUpdateIDs []uuid.UUID
		wantDeleteIDs []uuid.UUID
	}{
		{
			name:          "no changes",
			server:        nil,
			client:        nil,
			wantCreateIDs: nil,
			wantUpdateIDs: nil,
			wantDeleteIDs: nil,
		},
		{
			name:          "delete only",
			server:        nil,
			client:        []entities.EntryVersion{{ID: uuid1, Version: 1}},
			wantCreateIDs: nil,
			wantUpdateIDs: nil,
			wantDeleteIDs: []uuid.UUID{uuid1},
		},
		{
			name:          "create only",
			server:        []entities.EntryVersion{{ID: uuid1, Version: 1}},
			client:        nil,
			wantCreateIDs: []uuid.UUID{uuid1},
			wantUpdateIDs: nil,
			wantDeleteIDs: nil,
		},
		{
			name:          "update only",
			server:        []entities.EntryVersion{{ID: uuid1, Version: 2}},
			client:        []entities.EntryVersion{{ID: uuid1, Version: 1}},
			wantCreateIDs: nil,
			wantUpdateIDs: []uuid.UUID{uuid1},
			wantDeleteIDs: nil,
		},
		{
			name: "mixed changes",
			server: []entities.EntryVersion{
				{ID: uuid1, Version: 1},
				{ID: uuid2, Version: 1},
				{ID: uuid3, Version: 1},
			},
			client: []entities.EntryVersion{
				{ID: uuid2, Version: 1},
				{ID: uuid3, Version: 2},
				{ID: uuid4, Version: 1},
			},
			wantCreateIDs: []uuid.UUID{uuid1},
			wantUpdateIDs: []uuid.UUID{uuid3},
			wantDeleteIDs: []uuid.UUID{uuid4},
		},
	}

	merger := &Entry{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createIDs, updateIDs, deleteIDs, err := merger.GetDiff(
				context.Background(),
				tt.server,
				tt.client,
			)
			require.NoError(t, err, "no error expected")
			assert.ElementsMatch(t, tt.wantCreateIDs, createIDs)
			assert.ElementsMatch(t, tt.wantUpdateIDs, updateIDs)
			assert.ElementsMatch(t, tt.wantDeleteIDs, deleteIDs)
		})
	}
}
