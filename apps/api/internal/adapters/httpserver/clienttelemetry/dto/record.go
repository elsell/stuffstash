package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type Measurement struct {
	Platform   string  `json:"platform" enum:"ios,android,web"`
	Operation  string  `json:"operation" enum:"request,image"`
	Surface    string  `json:"surface" enum:"application,home,list,detail,gallery,fullscreen,upload"`
	Variant    string  `json:"variant" enum:"none,small,medium,large,original"`
	Outcome    string  `json:"outcome" enum:"success,failure,cancelled"`
	DurationMS float64 `json:"durationMs" minimum:"0" maximum:"60000"`
}
type RecordInput struct {
	Authorization string `header:"Authorization"`
	Body          struct {
		Measurements []Measurement `json:"measurements" minItems:"1" maxItems:"50"`
	}
}
type Accepted struct {
	Count int `json:"accepted"`
}
type RecordOutput struct {
	Body shared.SuccessEnvelope[Accepted]
}
