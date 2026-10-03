package brotherql

import (
	"bytes"
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/fakes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"
)

func label(t *testing.T) printing.Label {
	t.Helper()
	im := image.NewGray(image.Rect(0, 0, 306, 991))
	for i := range im.Pix {
		im.Pix[i] = 255
	}
	im.SetGray(0, 0, color.Gray{Y: 0})
	im.SetGray(305, 0, color.Gray{Y: 0})
	var data bytes.Buffer
	if err := png.Encode(&data, im); err != nil {
		t.Fatal(err)
	}
	return printing.Label{AttemptID: "attempt", Copy: 1, PNG: data.Bytes(), Media: Media()}
}
func TestCopyRequiresPhysicalCompletionAndCorrectPinPlacement(t *testing.T) {
	device := fakes.NewPrinterTransport()
	device.FragmentSize = 7
	connection := NewConnection(device)
	ctx := context.Background()
	state, err := connection.Readiness(ctx)
	if err != nil || state.State != printing.Ready {
		t.Fatal(state, err)
	}
	submission, err := connection.Submit(ctx, label(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(device.Rows) != 991 || device.Rows[0][0]&(1<<1) == 0 || device.Rows[0][38]&1 == 0 {
		t.Fatal("raster edge pixels did not land at pins6 and311")
	}
	for rowIndex, row := range device.Rows {
		for index, value := range row {
			if rowIndex == 0 && (index == 0 || index == 38) {
				continue
			}
			if value != 0 {
				t.Fatal("unexpected ink outside supplied pixels")
			}
		}
	}
	if _, err := connection.Submit(ctx, label(t)); err == nil {
		t.Fatal("allowed concurrent copy")
	}
	for i := 0; i < 3; i++ {
		result, err := connection.Observe(ctx, submission)
		if err != nil {
			t.Fatal(err)
		}
		if (i < 2 && result.Outcome != printing.Pending) || (i == 2 && result.Outcome != printing.Completed) {
			t.Fatalf("unexpected physical completion %#v", result)
		}
	}
	if result, err := connection.Observe(ctx, submission); err != nil || result.Outcome != printing.Completed {
		t.Fatal("confirmed completion evidence was lost", result, err)
	}
	if device.PhysicalLabels != 1 {
		t.Fatal("wrong physical output count")
	}
}
func TestUnavailableAndUncertainOutputAreNotRetried(t *testing.T) {
	for _, scenario := range []string{"paper", "partial", "lost-completion", "invalid-size"} {
		t.Run(scenario, func(t *testing.T) {
			device := fakes.NewPrinterTransport()
			connection := NewConnection(device)
			ctx := context.Background()
			if scenario == "paper" {
				device.Paper = false
			}
			state, err := connection.Readiness(ctx)
			if err != nil {
				t.Fatal(err)
			}
			input := label(t)
			switch scenario {
			case "paper":
				if state.State != printing.NeedsAttention {
					t.Fatal(state)
				}
			case "partial":
				device.FailAfterBytes = 1000
			case "lost-completion":
				device.DropCompletion = true
			case "invalid-size":
				input.Media.RasterWidth++
			}
			submission, err := connection.Submit(ctx, input)
			if scenario == "paper" || scenario == "invalid-size" {
				var failure *printing.SubmissionError
				if !errors.As(err, &failure) || failure.Outcome != printing.NoOutput || device.StartedLabels != 0 {
					t.Fatal("unavailable or invalid label caused output", err)
				}
				return
			}
			if scenario == "partial" {
				var failure *printing.SubmissionError
				if !errors.As(err, &failure) || failure.Outcome != printing.Uncertain || device.StartedLabels != 1 {
					t.Fatal("partial output not uncertain", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				deadline, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
				defer cancel()
				for {
					result, err := connection.Observe(deadline, submission)
					if err != nil {
						t.Fatal(err)
					}
					if result.Outcome == printing.Uncertain {
						break
					}
				}
			}
			before := device.StartedLabels
			if _, err := connection.Submit(ctx, input); err == nil || device.StartedLabels != before {
				t.Fatal("uncertain attempt replayed")
			}
		})
	}
}

func TestIdleConfirmationDoesNotSubmitAndRejectsActiveDevice(t *testing.T) {
	ctx := context.Background()
	device := fakes.NewPrinterTransport()
	connection := NewConnection(device)
	if err := connection.ConfirmIdle(ctx); err != nil {
		t.Fatal(err)
	}
	if device.PhysicalLabels != 0 || len(device.Rows) != 0 {
		t.Fatal("idle check printed")
	}
	if _, err := connection.Submit(ctx, label(t)); err != nil {
		t.Fatal(err)
	}
	if connection.ConfirmIdle(ctx) == nil {
		t.Fatal("active submission confirmed idle")
	}
}
