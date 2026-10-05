package inputfiles

import "io"

func prepareStdin(reader io.Reader) (io.Reader, func(), error) { return reader, func() {}, nil }
