package httpserver

import (
	"net/http/httptest"
	"testing"
)

type accessResponse struct {
	Relationship string   `json:"relationship"`
	Permissions  []string `json:"permissions"`
}

type myTenantResponse struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	LifecycleState string         `json:"lifecycleState"`
	Access         accessResponse `json:"access"`
}

type myTenantListBody struct {
	Data []myTenantResponse `json:"data"`
	Meta responseMeta       `json:"meta"`
}

func decodeMyTenantList(t *testing.T, response *httptest.ResponseRecorder) myTenantListBody {
	t.Helper()

	var body myTenantListBody
	decodeBody(t, response, &body)
	return body
}

func accessContainsPermission(permissions []string, expected string) bool {
	for _, permission := range permissions {
		if permission == expected {
			return true
		}
	}
	return false
}
