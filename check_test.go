package main

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_check(t *testing.T) {
	require := require.New(t)
	fs := afero.NewMemMapFs()
	require.NoError(fs.Mkdir("/test", 0644))
	require.NoError(afero.WriteFile(fs, "/test/[8AB2DCE2].txt", []byte("test1"), 0644))
	require.NoError(afero.WriteFile(fs, "/test/[00000000].txt", []byte("test2"), 0644))
	err := check(fs, "/test", false)
	assert.NoError(t, err)
}

func Test_checkCRC(t *testing.T) {
	type args struct {
		dir    string
		file   string
		update bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Example", args{"", "test[8AB2DCE2].txt", false}, false},
		{"Update", args{"", "test[00000000].txt", true}, false},
	}
	for _, tt := range tests {
		fs := afero.NewMemMapFs()
		err := afero.WriteFile(fs, "test[8AB2DCE2].txt", []byte("test1"), 0644)
		require.NoError(t, err)
		err = afero.WriteFile(fs, "test[00000000].txt", []byte("test2"), 0644)
		require.NoError(t, err)
		t.Run(tt.name, func(t *testing.T) {
			fi, err := fs.Stat(tt.args.file)
			assert.NoError(t, err)
			if err := checkCRC(fs, tt.args.dir, fi, tt.args.update); (err != nil) != tt.wantErr {
				t.Errorf("checkCRC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_extractHash(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"Extract", args{"[00000000].txt"}, []byte{0, 0, 0, 0}, false},
		{"NotFound", args{"[].txt"}, nil, false},
		{"NotHex", args{"[0000000Z].txt"}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractHash(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateHash(t *testing.T) {
	type args struct {
		name    string
		content string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"Example", args{"test.txt", "test"}, []byte{216, 127, 126, 12}, false},
		{"NoFile", args{"a", ""}, nil, true},
		{"NoContent", args{"test.txt", ""}, []byte{0, 0, 0, 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			err := afero.WriteFile(fs, "test.txt", []byte(tt.args.content), 0644)
			require.NoError(t, err)
			got, err := calculateHash(fs, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_renameFileHash(t *testing.T) {
	type args struct {
		dir          string
		file         string
		crcHashBytes []byte
		crcCalcBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Example",
			args{"/", "[00000000].txt", []byte{0, 0, 0, 0}, []byte{1, 1, 1, 1}},
			"[01010101].txt",
			false,
		},
	}
	for _, tt := range tests {
		fs := afero.NewMemMapFs()
		f, err := fs.Create(filepath.Join(tt.args.dir, tt.args.file))
		require.NoError(t, err)
		fi, err := f.Stat()
		require.NoError(t, err)
		t.Run(tt.name, func(t *testing.T) {
			if err := renameFileHash(fs, tt.args.dir, fi, tt.args.crcHashBytes, tt.args.crcCalcBytes); (err != nil) != tt.wantErr {
				t.Errorf("renameFileHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			exists, err := afero.Exists(fs, filepath.Join(tt.args.dir, tt.want))
			assert.NoError(t, err)
			assert.True(t, exists, tt.want+" doesn't exist")
		})
	}
}
