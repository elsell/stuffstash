package ports

import (
	"math"
	"slices"
)

type ClientPlatform string

const (
	ClientPlatformIos     ClientPlatform = "ios"
	ClientPlatformAndroid ClientPlatform = "android"
	ClientPlatformWeb     ClientPlatform = "web"
)

type ClientOperation string

const (
	ClientOperationRequest ClientOperation = "request"
	ClientOperationImage   ClientOperation = "image"
)

type ClientSurface string

const (
	ClientSurfaceHome       ClientSurface = "home"
	ClientSurfaceList       ClientSurface = "list"
	ClientSurfaceDetail     ClientSurface = "detail"
	ClientSurfaceGallery    ClientSurface = "gallery"
	ClientSurfaceFullscreen ClientSurface = "fullscreen"
	ClientSurfaceUpload     ClientSurface = "upload"
)

type ClientVariant string

const (
	ClientVariantNone     ClientVariant = "none"
	ClientVariantSmall    ClientVariant = "small"
	ClientVariantMedium   ClientVariant = "medium"
	ClientVariantLarge    ClientVariant = "large"
	ClientVariantOriginal ClientVariant = "original"
)

type ClientOutcome string

const (
	ClientOutcomeSuccess   ClientOutcome = "success"
	ClientOutcomeFailure   ClientOutcome = "failure"
	ClientOutcomeCancelled ClientOutcome = "cancelled"
)

type ClientMeasurement struct {
	Platform   ClientPlatform
	Operation  ClientOperation
	Surface    ClientSurface
	Variant    ClientVariant
	Outcome    ClientOutcome
	DurationMS float64
}

func (m ClientMeasurement) Valid() bool {
	return !math.IsNaN(m.DurationMS) && !math.IsInf(m.DurationMS, 0) && m.DurationMS >= 0 && m.DurationMS <= 60000 &&
		slices.Contains([]ClientPlatform{ClientPlatformIos, ClientPlatformAndroid, ClientPlatformWeb}, m.Platform) &&
		slices.Contains([]ClientOperation{ClientOperationRequest, ClientOperationImage}, m.Operation) &&
		slices.Contains([]ClientSurface{ClientSurfaceHome, ClientSurfaceList, ClientSurfaceDetail, ClientSurfaceGallery, ClientSurfaceFullscreen, ClientSurfaceUpload}, m.Surface) &&
		slices.Contains([]ClientVariant{ClientVariantNone, ClientVariantSmall, ClientVariantMedium, ClientVariantLarge, ClientVariantOriginal}, m.Variant) &&
		slices.Contains([]ClientOutcome{ClientOutcomeSuccess, ClientOutcomeFailure, ClientOutcomeCancelled}, m.Outcome)
}
