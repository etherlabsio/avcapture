package recorder

import "time"

type FFmpeg struct {
	Params  [][]string `json:"params"`
	Options [][]string `json:"options"`
}
type Chrome struct {
	URL     string   `json:"url"`
	Options []string `json:"options"`
}

// StartRecordingRequest is the payload being received by recorder as part of start_recording
type StartRecordingRequest struct {
	FFmpeg `json:"ffmpeg"`
	Chrome `json:"chrome"`
}
type StartRecordingResponse struct {
	StartTime time.Time `json:"start_time"`
}
type StopRecordingResponse struct {
	StopTime time.Time `json:"stop_time"`
}
