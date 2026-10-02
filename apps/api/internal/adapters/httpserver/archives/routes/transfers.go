package routes

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/archives/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/archives/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
)

// Register streaming operations directly through Huma's adapter so archive bytes
// never pass through the JSON/body-buffer decoder. The same API middleware and
// explicit OpenAPI contract still apply.
func registerTransfers(api huma.API, application app.App, service *dataportability.ArchiveService, timeout time.Duration) {
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	response := api.OpenAPI().Components.Schemas.Schema(reflect.TypeFor[shared.SuccessEnvelope[dto.Job]](), true, "ArchiveJobEnvelope")
	upload := huma.Operation{OperationID: "upload-archive-restore", Method: http.MethodPost, Path: "/tenants/{tenantId}/archive-restores", Tags: []string{"archives"}, DefaultStatus: 201, Parameters: []*huma.Param{{Name: "tenantId", In: "path", Required: true, Schema: &huma.Schema{Type: "string"}}, {Name: "Idempotency-Key", In: "header", Required: true, Schema: &huma.Schema{Type: "string"}}}, RequestBody: &huma.RequestBody{Required: true, Content: map[string]*huma.MediaType{"application/zip": {Schema: &huma.Schema{Type: "string", Format: "binary"}}}}, Responses: map[string]*huma.Response{"201": {Description: "Restore uploaded for validation", Content: map[string]*huma.MediaType{"application/json": {Schema: response}}}}}
	addTransferErrors(api, &upload)
	shared.SecuredOperation(&upload)
	api.OpenAPI().AddOperation(&upload)
	api.Adapter().Handle(&upload, api.Middlewares().Handler(func(c huma.Context) {
		access, err := authenticate(c.Context(), application, service, dto.Access{Authorization: c.Header("Authorization"), TenantID: c.Param("tenantId")})
		if err != nil {
			writeTransferError(api, c, err)
			return
		}
		key := strings.TrimSpace(c.Header("Idempotency-Key"))
		if key == "" || len(key) > 200 {
			writeTransferError(api, c, huma.Error400BadRequest("Idempotency-Key is required and must be at most 200 characters."))
			return
		}
		contentType, _, err := mime.ParseMediaType(c.Header("Content-Type"))
		if err != nil || contentType != "application/zip" {
			writeTransferError(api, c, huma.Error415UnsupportedMediaType("Upload a Stuff Stash ZIP archive."))
			return
		}
		ctx, cancel := context.WithTimeout(c.Context(), timeout)
		defer cancel()
		request, writer := humago.Unwrap(c)
		controller := http.NewResponseController(writer)
		_ = controller.SetReadDeadline(time.Now().Add(timeout))
		_ = controller.SetWriteDeadline(time.Now().Add(timeout))
		job, err := service.UploadRestore(ctx, access, key, request.Body)
		if err != nil {
			writeTransferError(api, c, shared.ToHumaError(err))
			return
		}
		c.SetHeader("Content-Type", "application/json")
		c.SetHeader("Cache-Control", "private, no-store")
		c.SetStatus(201)
		_ = json.NewEncoder(c.BodyWriter()).Encode(shared.SuccessEnvelope[dto.Job]{Data: mapper.Job(job), Meta: shared.Meta{TenantID: access.TenantID.String()}})
	}))
	download := huma.Operation{OperationID: "download-inventory-archive", Method: http.MethodGet, Path: jobsPath + "/{jobId}/content", Tags: []string{"archives"}, Parameters: []*huma.Param{{Name: "tenantId", In: "path", Required: true, Schema: &huma.Schema{Type: "string"}}, {Name: "jobId", In: "path", Required: true, Schema: &huma.Schema{Type: "string"}}, {Name: "inventoryId", In: "query", Schema: &huma.Schema{Type: "string"}}}, Responses: map[string]*huma.Response{"200": {Description: "Portable inventory ZIP", Content: map[string]*huma.MediaType{"application/zip": {Schema: &huma.Schema{Type: "string", Format: "binary"}}}}}}
	addTransferErrors(api, &download)
	shared.SecuredOperation(&download)
	api.OpenAPI().AddOperation(&download)
	api.Adapter().Handle(&download, api.Middlewares().Handler(func(c huma.Context) {
		access, err := authenticate(c.Context(), application, service, dto.Access{Authorization: c.Header("Authorization"), TenantID: c.Param("tenantId"), InventoryID: c.Query("inventoryId")})
		if err != nil {
			writeTransferError(api, c, err)
			return
		}
		ctx, cancel := context.WithTimeout(c.Context(), timeout)
		defer cancel()
		_, writer := humago.Unwrap(c)
		_ = http.NewResponseController(writer).SetWriteDeadline(time.Now().Add(timeout))
		stream, size, err := service.Download(ctx, access, c.Param("jobId"))
		if err != nil {
			writeTransferError(api, c, shared.ToHumaError(err))
			return
		}
		defer stream.Close()
		c.SetHeader("Content-Type", "application/zip")
		c.SetHeader("Content-Disposition", "attachment; filename=\"stuff-stash-inventory.zip\"")
		c.SetHeader("Cache-Control", "private, no-store")
		c.SetHeader("Content-Length", strconv.FormatInt(size, 10))
		c.SetStatus(200)
		// Content-Length makes a partial transfer detectable by recipients. Do not
		// append a JSON error once a ZIP response has started.
		if _, err = io.CopyN(c.BodyWriter(), stream, size); err != nil {
			panic(http.ErrAbortHandler)
		}
	}))
}
func writeTransferError(api huma.API, c huma.Context, err error) {
	var status huma.StatusError
	if errors.As(err, &status) {
		_ = huma.WriteErr(api, c, status.GetStatus(), status.Error())
		return
	}
	_ = huma.WriteErr(api, c, 500, "Internal server error.")
}

func addTransferErrors(api huma.API, operation *huma.Operation) {
	schema := api.OpenAPI().Components.Schemas.Schema(reflect.TypeFor[shared.ErrorEnvelope](), true, "ErrorEnvelope")
	for _, code := range []int{400, 401, 403, 404, 409, 413, 415, 500, 503} {
		operation.Responses[strconv.Itoa(code)] = &huma.Response{Description: http.StatusText(code), Content: map[string]*huma.MediaType{"application/json": {Schema: schema}}}
	}
}
