package marshal

import (
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/core"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	emptyDataErr = func(t require.TestingT, err error, args ...interface{}) {
		require.ErrorIs(t, err, entities.ErrEntryDataEmpty, "want data empty error")
	}
)

func TestEntryMarshaler(t *testing.T) {
	tests := []struct {
		name string
		typ  core.EntryType
		data entities.EntryData
	}{
		{
			name: "password",
			typ:  core.EntryTypePassword,
			data: entities.EntryDataPassword{
				Login:    "login",
				Password: "password",
			},
		},
		{
			name: "card",
			typ:  core.EntryTypeCard,
			data: entities.EntryDataCard{
				Number:  "4012000033330026",
				Expires: "11/2026",
				Cvc:     "371",
				Owner:   "Robert",
			},
		},
		{
			name: "note",
			typ:  core.EntryTypeNote,
			data: entities.EntryDataNote("note"),
		},
		{
			name: "binary",
			typ:  core.EntryTypeBinary,
			data: entities.EntryDataBinary("binary"),
		},
	}

	sut := EntryMarshaler{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sut.Marshal(tt.data)
			require.NoError(t, err, "no error expected when marshaling entry data")
			got2, err := sut.Unmarshal(tt.typ, got)
			require.NoError(t, err, "no error expected when unmarshaling entry data")
			require.Equal(t, tt.data, got2, "original and unmarshaled data should be equal")
		})
	}
}

func TestEntryMarshaler_Marshal(t *testing.T) {

	tests := []struct {
		name    string
		typ     core.EntryType
		data    entities.EntryData
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "invalid: nil data",
			data:    nil,
			wantErr: emptyDataErr,
		},
		{
			name: "invalid: password pointer",
			data: &entities.EntryDataPassword{},
			wantErr: func(t require.TestingT, err error, args ...interface{}) {
				require.ErrorIs(t, err, entities.ErrEntryDataTypeInvalid, "want data empty error")
			},
		},
		{
			name:    "valid : empty password",
			data:    entities.EntryDataPassword{},
			wantErr: require.NoError,
		},
		{
			name:    "valid: empty card",
			data:    entities.EntryDataCard{},
			wantErr: require.NoError,
		},
		{
			name:    "invalid: empty note",
			data:    entities.EntryDataNote(""),
			wantErr: emptyDataErr,
		},
		{
			name:    "invalid: empty binary",
			data:    entities.EntryDataBinary(""),
			wantErr: emptyDataErr,
		},
	}

	sut := EntryMarshaler{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sut.Marshal(tt.data)
			tt.wantErr(t, err)
		})
	}
}

func TestEntryMarshaler_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		typ     core.EntryType
		data    []byte
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "invalid: nil password data",
			typ:     core.EntryTypePassword,
			data:    nil,
			wantErr: emptyDataErr,
		},
		{
			name:    "invalid: empty password data",
			typ:     core.EntryTypePassword,
			data:    []byte("{}"),
			wantErr: require.NoError,
		},
		{
			name:    "invalid: empty card data",
			typ:     core.EntryTypeCard,
			data:    []byte("{}"),
			wantErr: require.NoError,
		},
		{
			name:    "invalid: empty note data",
			typ:     core.EntryTypeNote,
			data:    []byte(""),
			wantErr: emptyDataErr,
		},
		{
			name:    "invalid: empty binary data",
			typ:     core.EntryTypeBinary,
			data:    []byte(""),
			wantErr: emptyDataErr,
		},
	}
	sut := EntryMarshaler{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sut.Unmarshal(tt.typ, tt.data)
			tt.wantErr(t, err)
		})
	}
}
