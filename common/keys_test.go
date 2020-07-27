package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKeyToUint(t *testing.T) {
	a := assert.New(t)
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want uint
	}{
		{name: "A", args: args{key: "A"}, want: 0x41},
		{name: "0", args: args{key: "0"}, want: 0x30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a.Equal(tt.want, KeyToUint(tt.args.key))
		})
	}
}

func TestKeyvalName(t *testing.T) {
	a := assert.New(t)
	type args struct {
		keyval uint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "A", args: args{keyval: 0x41}, want: "A"},
		{name: "A", args: args{keyval: 0x30}, want: "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a.Equal(tt.want, KeyvalName(tt.args.keyval))
		})
	}
}
