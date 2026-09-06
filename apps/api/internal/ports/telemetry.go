package ports

import "context"

// Operation names are bounded application vocabulary, never resource identifiers.
type Operation string

const (
	OperationModelImagePrepare Operation = "media.model_image.prepare"
	OperationUploadInitiate    Operation = "media.upload.initiate"
	OperationThumbnailGenerate Operation = "media.thumbnail.generate"
	OperationBlobRead          Operation = "media.blob.read"
	OperationBlobWrite         Operation = "media.blob.write"
	OperationBlobDelete        Operation = "media.blob.delete"
	OperationUploadVerify      Operation = "media.upload.verify"
	OperationHTTP              Operation = "http.request"
	OperationOther             Operation = "other"
)

func (o Operation) Bounded() Operation {
	switch o {
	case OperationModelImagePrepare, OperationUploadInitiate, OperationThumbnailGenerate, OperationBlobRead, OperationBlobWrite, OperationBlobDelete, OperationUploadVerify, OperationHTTP:
		return o
	default:
		return OperationOther
	}
}

// Telemetry measures an operation without coupling application code to an SDK.
// Completion must be safe to call more than once, including from deferred cleanup.
type Telemetry interface {
	Start(context.Context, Operation) (context.Context, func(error))
}

type NoopTelemetry struct{}

func (NoopTelemetry) Start(ctx context.Context, _ Operation) (context.Context, func(error)) {
	return ctx, func(error) {}
}
