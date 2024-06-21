package mapper

import (
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"github.com/dlomanov/gophkeeper/internal/core"
)

type EntryMapper struct{}

func (EntryMapper) ToEntityType(t pb.EntryType) core.EntryType {
	switch t {
	case pb.EntryType_ENTRY_TYPE_UNSPECIFIED:
		return core.EntryTypeUnspecified
	case pb.EntryType_ENTRY_TYPE_PASSWORD:
		return core.EntryTypePassword
	case pb.EntryType_ENTRY_TYPE_NOTE:
		return core.EntryTypeNote
	case pb.EntryType_ENTRY_TYPE_CARD:
		return core.EntryTypeCard
	case pb.EntryType_ENTRY_TYPE_BINARY:
		return core.EntryTypeBinary
	default:
		return core.EntryTypeUnspecified
	}
}

func (EntryMapper) ToAPIType(t core.EntryType) pb.EntryType {
	switch t {
	case core.EntryTypeUnspecified:
		return pb.EntryType_ENTRY_TYPE_UNSPECIFIED
	case core.EntryTypePassword:
		return pb.EntryType_ENTRY_TYPE_PASSWORD
	case core.EntryTypeNote:
		return pb.EntryType_ENTRY_TYPE_NOTE
	case core.EntryTypeCard:
		return pb.EntryType_ENTRY_TYPE_CARD
	case core.EntryTypeBinary:
		return pb.EntryType_ENTRY_TYPE_BINARY
	default:
		return pb.EntryType_ENTRY_TYPE_UNSPECIFIED
	}
}
