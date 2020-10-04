package apitype

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScaleToFit(t *testing.T) {
	a := assert.New(t)
	type args struct {
		sourceWidth  int
		sourceHeight int
		targetWidth  int
		targetHeight int
	}
	tests := []struct {
		name   string
		args   args
		width  int
		height int
	}{
		{name: "100x100->100x100", args: args{sourceWidth: 100, sourceHeight: 100, targetWidth: 100, targetHeight: 100}, width: 100, height: 100},
		// Downscale
		{name: "200x200->100x100", args: args{sourceWidth: 200, sourceHeight: 200, targetWidth: 100, targetHeight: 100}, width: 100, height: 100},
		{name: "400x300->100x100", args: args{sourceWidth: 400, sourceHeight: 300, targetWidth: 100, targetHeight: 100}, width: 100, height: 75},
		{name: "400x300->100x50", args: args{sourceWidth: 400, sourceHeight: 300, targetWidth: 100, targetHeight: 50}, width: 66, height: 50},
		{name: "300x400->100x100", args: args{sourceWidth: 300, sourceHeight: 400, targetWidth: 100, targetHeight: 100}, width: 75, height: 100},
		{name: "300x400->100x50", args: args{sourceWidth: 300, sourceHeight: 400, targetWidth: 100, targetHeight: 50}, width: 37, height: 50},
		// Upscale
		{name: "100x100->200x200", args: args{sourceWidth: 100, sourceHeight: 100, targetWidth: 200, targetHeight: 200}, width: 200, height: 200},
		{name: "40x30  ->400x400", args: args{sourceWidth: 40, sourceHeight: 30, targetWidth: 400, targetHeight: 400}, width: 400, height: 300},
		{name: "30x40  ->400x400", args: args{sourceWidth: 30, sourceHeight: 40, targetWidth: 400, targetHeight: 400}, width: 300, height: 400},
		{name: "30x40  ->400x100", args: args{sourceWidth: 30, sourceHeight: 40, targetWidth: 400, targetHeight: 100}, width: 75, height: 100},
		{name: "40x30  ->400x100", args: args{sourceWidth: 40, sourceHeight: 30, targetWidth: 400, targetHeight: 100}, width: 133, height: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := ScaleToFit(tt.args.sourceWidth, tt.args.sourceHeight, tt.args.targetWidth, tt.args.targetHeight)
			a.Equal(tt.width, w)
			a.Equal(tt.height, h)
		})
	}
}

func TestSizeOf(t *testing.T) {
	a := assert.New(t)
	type args struct {
		width  int
		height int
	}
	tests := []struct {
		name          string
		args          args
		width, height int
	}{
		{name: "Size", args: args{width: 200, height: 100}, width: 200, height: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SizeOf(tt.args.width, tt.args.height)
			a.Equal(tt.width, got.GetWidth())
			a.Equal(tt.height, got.GetHeight())
		})
	}
}
