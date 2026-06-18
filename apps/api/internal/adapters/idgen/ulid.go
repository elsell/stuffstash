package idgen

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

type ULIDGenerator struct{}

func NewULIDGenerator() ULIDGenerator {
	return ULIDGenerator{}
}

func (ULIDGenerator) NewID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now().UTC()), rand.Reader).String()
}
