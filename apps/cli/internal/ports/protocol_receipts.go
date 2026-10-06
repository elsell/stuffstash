package ports

import "context"

// ProtocolReceiptResult admits only reviewed, nonsecret response projections.
type ProtocolReceiptResult interface{ protocolReceiptResult() }
type ProtocolReceipt struct {
	Operation string                `json:"operation"`
	Result    ProtocolReceiptResult `json:"result"`
}

// Record cannot fail a successful mutation or trigger its retry.
type ProtocolReceipts interface {
	Record(context.Context, ProtocolReceipt)
}
type WorkerAttemptReceipt Result[*ConsumerAttempt]

func (WorkerAttemptReceipt) protocolReceiptResult() {}

type WorkerConnectorReceipt Result[PrintConnector]

func (WorkerConnectorReceipt) protocolReceiptResult() {}

type WorkerAcknowledgementReceipt Result[struct{}]

func (WorkerAcknowledgementReceipt) protocolReceiptResult() {}

type VerifiedPrintArtifactReceipt struct {
	AttemptID    string `json:"attemptId"`
	SHA256       string `json:"sha256"`
	ContentType  string `json:"contentType"`
	ByteLength   int64  `json:"byteLength"`
	WidthPixels  int    `json:"widthPixels"`
	HeightPixels int    `json:"heightPixels"`
}

func (VerifiedPrintArtifactReceipt) protocolReceiptResult() {}
