package printworker_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/png"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/brotherql"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/fakes"
)

func TestWorkerAndBrotherProtocolCompleteCopiesWithoutReplayAfterLostReply(t *testing.T) {
	w, a, j, _ := fixture(t)
	media := brotherql.Media()
	var body bytes.Buffer
	if err := png.Encode(&body, image.NewGray(image.Rect(0, 0, media.RasterWidth, media.RasterHeight))); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body.Bytes())
	a.bytes = body.Bytes()
	a.claim.Media = media
	a.claim.Artifact = printing.Artifact{SHA256: hex.EncodeToString(sum[:]), ContentType: "image/png", ByteLength: int64(body.Len()), Width: media.RasterWidth, Height: media.RasterHeight}
	a.dropOutcome = true
	w.Config.Media = media
	hardware := fakes.NewPrinterTransport()
	connection := brotherql.NewConnection(hardware)
	defer connection.Close()
	if err := w.Step(context.Background(), j, connection); err == nil {
		t.Fatal("lost acknowledgement hidden")
	}
	if hardware.PhysicalLabels != 2 || hardware.StartedLabels != 2 {
		t.Fatalf("copies: completed %d started %d", hardware.PhysicalLabels, hardware.StartedLabels)
	}
	w.Config.SessionID = "replacement-process"
	if err := w.Step(context.Background(), j, connection); err != nil {
		t.Fatal(err)
	}
	if hardware.PhysicalLabels != 2 || j.record != nil {
		t.Fatal("recovery replayed physical output")
	}
}
