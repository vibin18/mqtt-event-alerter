package messengers

import (
	"io"
)

type Alerter interface {
	SendTextAlert(string)
	SendPictureAlert(label, camera, eventId, contentTime string, picture io.ReadCloser)
}
