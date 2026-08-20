//go:build !bench
// +build !bench

package hw10programoptimization

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat(t *testing.T) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`

	t.Run("find 'com'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"browsecat.com": 2,
			"linktype.com":  1,
		}, result)
	})

	t.Run("find 'gov'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "gov")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"browsedrive.gov": 1}, result)
	})

	t.Run("find 'unknown'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "unknown")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})
}

func TestGetDomainStat_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		domain  string
		want    DomainStat
		wantErr bool
	}{
		{
			name:   "empty reader",
			input:  "",
			domain: "com",
			want:   DomainStat{},
		},
		{
			name: "skips empty lines",
			input: `{"Email":"a@foo.com"}

{"Email":"b@bar.com"}
`,
			domain: "com",
			want: DomainStat{
				"foo.com": 1,
				"bar.com": 1,
			},
		},
		{
			name:   "last line without newline",
			input:  "{\"Email\":\"a@foo.com\"}\n{\"Email\":\"b@bar.com\"}",
			domain: "com",
			want: DomainStat{
				"foo.com": 1,
				"bar.com": 1,
			},
		},
		{
			name:   "counts same domain",
			input:  "{\"Email\":\"a@foo.com\"}\n{\"Email\":\"b@foo.com\"}\n",
			domain: "com",
			want:   DomainStat{"foo.com": 2},
		},
		{
			name:   "normalizes domain case",
			input:  `{"Email":"User@OtUs.ru"}`,
			domain: "ru",
			want:   DomainStat{"otus.ru": 1},
		},
		{
			name:   "does not match domain substring without dot",
			input:  `{"Email":"User@company.gov"}`,
			domain: "com",
			want:   DomainStat{},
		},
		{
			name:   "skips email without @",
			input:  `{"Email":"not-an-email.com"}`,
			domain: "com",
			want:   DomainStat{},
		},
		{
			name:    "invalid json",
			input:   `{"Email":`,
			domain:  "com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetDomainStat(bytes.NewBufferString(tt.input), tt.domain)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, result)
		})
	}
}

type failReader struct{}

func (failReader) Read(_ []byte) (int, error) {
	return 0, errors.New("some broken reader")
}

func TestGetDomainStat_ReadError(t *testing.T) {
	result, err := GetDomainStat(failReader{}, "com")
	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorContains(t, err, "some broken reader")
}
