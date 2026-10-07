package httpapi

import (
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func binaryContent(response *http.Response, err error) (ports.BinaryContent, error) {
	if err != nil {
		return ports.BinaryContent{}, ports.Failure("network", "The CLI cannot download the file. Make sure that the server is available. Then try again.")
	}
	if response.StatusCode != http.StatusOK {
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			_, err = read[any](response, nil)
			return ports.BinaryContent{}, err
		}
		response.Body.Close()
		return ports.BinaryContent{}, ports.Failure("protocol", "The server did not return a complete file. Try the download again.")
	}
	return ports.BinaryContent{Body: response.Body, ContentType: response.Header.Get("Content-Type"), ContentDisposition: response.Header.Get("Content-Disposition"), ContentLength: response.ContentLength}, nil
}
