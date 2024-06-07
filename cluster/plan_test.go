package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getUniqStringList(t *testing.T) {
	type args struct {
		l []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"contain strings with only spaces",
			args{
				[]string{" ", "key1=value1", "   ", "key2=value2"},
			},
			[]string{"key1=value1", "key2=value2"},
		},
		{
			"contain strings with trailing or leading spaces",
			args{
				[]string{"  key1=value1", "key1=value1  ", "  key2=value2   "},
			},
			[]string{"key1=value1", "key2=value2"},
		},
		{
			"contain duplicated strings",
			args{
				[]string{"", "key1=value1", "key1=value1", "key2=value2"},
			},
			[]string{"key1=value1", "key2=value2"},
		},
		{
			"contain empty string",
			args{
				[]string{"", "key1=value1", "", "key2=value2"},
			},
			[]string{"key1=value1", "key2=value2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getUniqStringList(tt.args.l), "getUniqStringList(%v)", tt.args.l)
		})
	}
}
