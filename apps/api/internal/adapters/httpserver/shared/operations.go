package shared

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func SecuredOperation(operation *huma.Operation) {
	operation.Security = []map[string][]string{{"bearerAuth": {}}}
}

func CreatedOperation(operation *huma.Operation) {
	operation.DefaultStatus = http.StatusCreated
}
