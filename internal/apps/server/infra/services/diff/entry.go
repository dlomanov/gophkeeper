package diff

import (
	"context"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/google/uuid"
)

var _ usecases.EntryDiffer = (*Entry)(nil)

type Entry struct {
}

func NewEntry() *Entry {
	return &Entry{}
}

func (e Entry) GetDiff(
	_ context.Context,
	server []core.EntryVersion,
	clientVersions []core.EntryVersion,
) (
	createIDs []uuid.UUID,
	updateIDs []uuid.UUID,
	deleteIDs []uuid.UUID,
	err error) {
	var (
		serverMap = make(map[uuid.UUID]core.EntryVersion)
		clientMap = make(map[uuid.UUID]core.EntryVersion)
	)
	for _, v := range server {
		serverMap[v.ID] = v
	}
	for _, v := range clientVersions {
		clientMap[v.ID] = v
	}

	for _, s := range server {
		// server and client have the same entry
		if c, ok := clientMap[s.ID]; ok {
			// but different version
			if s.Version != c.Version {
				updateIDs = append(updateIDs, s.ID)
			}
			continue
		}
		// client does not have the entry
		createIDs = append(createIDs, s.ID)
	}

	for _, c := range clientVersions {
		// client have deleted entry
		if _, ok := serverMap[c.ID]; !ok {
			deleteIDs = append(deleteIDs, c.ID)
		}
	}

	return createIDs, updateIDs, deleteIDs, nil
}
