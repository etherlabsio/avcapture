package recorder

import (
	"context"

	"github.com/etherlabsio/errors"
)

func ValidationMiddleware(svc Service) Service {
	return validationMiddleware{next: svc}
}

type validationMiddleware struct {
	next Service
}

func (mw validationMiddleware) Start(ctx context.Context, req StartRecordingRequest) (resp StartRecordingResponse) {
	var op errors.Op = "Start"
	if "" == req.URL {
		resp.Err = errors.New("missing URL for capturing", op, errors.Invalid)
		return resp
	}
	return mw.next.Start(ctx, req)
}

func (mw validationMiddleware) Stop(ctx context.Context, req StopRecordingRequest) StopRecordingResponse {
	return mw.next.Stop(ctx, req)
}

func (mw validationMiddleware) Check(ctx context.Context) error {
	return mw.next.Check(ctx)
}

func (mw validationMiddleware) MuteRecordingAudio(ctx context.Context) error {
	return mw.next.MuteRecordingAudio(ctx)
}

func (mw validationMiddleware) UnmuteRecordingAudio(ctx context.Context) error {
	return mw.next.UnmuteRecordingAudio(ctx)
}
