package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type RegisterDeviceInput struct {
	ScopeInput
	Body RegisterDeviceBody
}
type RegisterDeviceBody struct {
	InstallationID string `json:"installationId" minLength:"1" maxLength:"128"`
	Transport      string `json:"transport" enum:"apns,fcm"`
	Token          string `json:"token" writeOnly:"true"`
	Revision       int64  `json:"revision" minimum:"0"`
}
type GetDeviceInput struct {
	ScopeInput
	InstallationID string `path:"installationId" maxLength:"128"`
}
type RevokeDeviceInput struct {
	ScopeInput
	DeviceID string `path:"deviceId"`
	Revision int64  `query:"revision" minimum:"1" required:"true"`
}
type DeviceResponse struct {
	ID             string `json:"id"`
	InstallationID string `json:"installationId"`
	Transport      string `json:"transport"`
	Revision       int64  `json:"revision"`
	Active         bool   `json:"active"`
}
type DeviceOutput struct {
	Body shared.SuccessEnvelope[DeviceResponse]
}
