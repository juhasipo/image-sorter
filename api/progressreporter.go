package api

type ProgressReporter interface {
	Update(name string, current int, total int, canCancel bool, modal bool)
	Error(error string, err error)
}

type SenderProgressReporter struct {
	sender Sender

	ProgressReporter
}

func NewSenderProgressReporter(sender Sender) ProgressReporter {
	return SenderProgressReporter{
		sender: sender,
	}
}

func (s SenderProgressReporter) Update(name string, current int, total int, canCancel bool, modal bool) {
	s.sender.SendCommandToTopic(ProcessStatusUpdated, &UpdateProgressCommand{
		Name:      name,
		Current:   current,
		Total:     total,
		CanCancel: canCancel,
		Modal:     modal,
	})
}

func (s SenderProgressReporter) Error(error string, err error) {
	s.sender.SendError(error, err)
}
