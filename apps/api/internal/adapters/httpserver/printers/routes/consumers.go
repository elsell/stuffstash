package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"strings"
)

func consumer(ctx context.Context, application app.App, authorization string) (printing.Connector, error) {
	if !application.PrintConnectorsConfigured() {
		return printing.Connector{}, huma.Error503ServiceUnavailable("Print connectors are unavailable")
	}
	scheme, token, ok := strings.Cut(authorization, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return printing.Connector{}, huma.Error401Unauthorized("Invalid connector credential")
	}
	c, err := application.PrintConnectors().AuthenticateConsumer(ctx, token)
	if err != nil {
		return printing.Connector{}, shared.ToHumaError(err)
	}
	return c, nil
}
func registerConsumerPrinters(api huma.API, application app.App) {
	huma.Get(api, "/print-consumer/printers", func(ctx context.Context, input *dto.ConsumerInput) (*dto.ConsumerPrintersOutput, error) {
		c, err := consumer(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		printers, err := application.PrintConnectors().ConsumerPrinters(ctx, c)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		result := make([]dto.ConsumerPrinter, 0, len(printers))
		for _, p := range printers {
			result = append(result, dto.ConsumerPrinter{Printer: mapper.Printer(p.Printer), DeviceID: p.DeviceID, BindingGeneration: p.BindingGeneration})
		}
		return &dto.ConsumerPrintersOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[[]dto.ConsumerPrinter]{Data: result}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Post(api, "/print-consumer/printer-reports", func(ctx context.Context, input *dto.PrinterReportInput) (*dto.PrinterReportOutput, error) {
		c, err := consumer(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		if err := application.PrintConnectors().ReportPrinter(ctx, c, printing.PrinterID(input.Body.PrinterID), printing.PrinterReadiness(input.Body.State), input.Body.Reason); err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.PrinterReportOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[struct{}]{Data: struct{}{}}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
}
