package marshal

import (
	"encoding/json"
	"fmt"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/core"
)

type EntryMarshaler struct{}

func (EntryMarshaler) Marshal(data entities.EntryData) (result []byte, err error) {
	if data == nil {
		return nil, fmt.Errorf("%w: data empty", entities.ErrEntryDataEmpty)
	}
	switch data := data.(type) {
	case entities.EntryDataPassword:
		if result, err = json.Marshal(data); err != nil {
			return nil, fmt.Errorf("entry_marshaler: failed to marshal password data: %w", err)
		}
	case entities.EntryDataNote:
		result = []byte(data)
	case entities.EntryDataCard:
		if result, err = json.Marshal(data); err != nil {
			return nil, fmt.Errorf("entry_marshaler: failed to marshal card data: %w", err)
		}
	case entities.EntryDataBinary:
		result = data
	default:
		return nil, fmt.Errorf("entry_marshaler: unknown type: %w", entities.ErrEntryDataTypeInvalid)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("entry_marshaler: %w", entities.ErrEntryDataEmpty)
	}
	return result, nil
}

func (EntryMarshaler) Unmarshal(typ core.EntryType, data []byte) (entities.EntryData, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("entry_marshaler: %w", entities.ErrEntryDataEmpty)
	}
	switch typ {
	case core.EntryTypePassword:
		var passData entities.EntryDataPassword
		if err := json.Unmarshal(data, &passData); err != nil {
			return nil, fmt.Errorf("entry_marshaler: failed to unmarshal password data: %w", err)
		}
		return passData, nil
	case core.EntryTypeNote:
		return entities.EntryDataNote(data), nil
	case core.EntryTypeCard:
		var cardData entities.EntryDataCard
		if err := json.Unmarshal(data, &cardData); err != nil {
			return nil, fmt.Errorf("entry_marshaler: failed to unmarshal card data: %w", err)
		}
		return cardData, nil
	case core.EntryTypeBinary:
		return entities.EntryDataBinary(data), nil
	default:
		return nil, fmt.Errorf("%w: %s", entities.ErrEntryTypeInvalid, typ)
	}
}
