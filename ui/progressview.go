package ui

import (
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/event"
)

type ProgressView struct {
	view        *gtk.Box
	progressbar *gtk.ProgressBar
	stopButton  *gtk.Button
}

func NewProgressView(builder *gtk.Builder, sender event.Sender) *ProgressView {
	progressView := &ProgressView{
		view:        GetObjectOrPanic(builder, "progress-view").(*gtk.Box),
		stopButton:  GetObjectOrPanic(builder, "stop-progress-button").(*gtk.Button),
		progressbar: GetObjectOrPanic(builder, "progress-bar").(*gtk.ProgressBar),
	}
	progressView.stopButton.Connect("clicked", func() {
		sender.SendToTopic(event.SimilarRequestStop)
	})

	return progressView
}

func (v *ProgressView) SetVisible(visible bool) {
	v.view.SetVisible(visible)
}

func (v *ProgressView) SetStatus(status int, total int) {
	statusText := fmt.Sprintf("Processed %d/%d", status, total)
	v.progressbar.SetText(statusText)
	v.progressbar.SetFraction(float64(status) / float64(total))
}
