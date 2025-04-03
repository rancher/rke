package cluster

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
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

func Test_getStringChecksum(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		version  string
		expected string
	}{
		{
			name:     "version greater than 1.31.6, use sha256",
			config:   "test-config",
			version:  "v1.32.0-rancher0",
			expected: fmt.Sprintf("%x", sha256.Sum256([]byte("test-config"))),
		},
		{
			name:     "version exactly 1.31.6, use sha256",
			config:   "test-config",
			version:  "v1.31.6-rancher0",
			expected: fmt.Sprintf("%x", sha256.Sum256([]byte("test-config"))),
		},
		{
			name:     "version exactly 1.31.0, use md5",
			config:   "test-config",
			version:  "v1.31.0-rancher0",
			expected: fmt.Sprintf("%x", md5.Sum([]byte("test-config"))),
		},
		{
			name:     "version less than 1.31, use md5",
			config:   "test-config",
			version:  "v1.30.0-rancher0",
			expected: fmt.Sprintf("%x", md5.Sum([]byte("test-config"))),
		},
		{
			name:     "empty config",
			config:   "",
			version:  "v1.32.0-rancher0",
			expected: fmt.Sprintf("%x", sha256.Sum256([]byte(""))),
		},
		{
			name:     "empty version",
			config:   "test-config",
			version:  "",
			expected: fmt.Sprintf("%x", md5.Sum([]byte("test-config"))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := getStringChecksum(tt.config, tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}
