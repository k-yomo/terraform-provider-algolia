package algoliautil

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/transport"
)

type DebugRequester struct {
	Client *http.Client
}

func NewDebugRequester() *DebugRequester {
	httpClient := transport.DefaultHTTPClient()
	httpClient.Transport = newDebugTransport(httpClient.Transport)
	return &DebugRequester{
		Client: httpClient,
	}
}

func (d *DebugRequester) Request(req *http.Request) (*http.Response, error) {
	return d.Client.Do(req)
}

// Code from below is basically copied from the following logging helper
//(need to copy to mask secrets)
// https://github.com/hashicorp/terraform-plugin-sdk/blob/45133e6e2aebbe0aca05427cbcd360f968979e98/helper/logging/transport.go#L12
type debugTransport struct {
	name      string
	transport http.RoundTripper
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqData, err := httputil.DumpRequestOut(req, true)
	if err == nil {
		log.Printf("[DEBUG] "+logReqMsg, t.name, prettyPrintJsonLines(reqData))
	} else {
		log.Printf("[ERROR] %s API Request error: %#v", t.name, err)
	}

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	respData, err := httputil.DumpResponse(resp, true)
	if err == nil {
		log.Printf("[DEBUG] "+logRespMsg, t.name, prettyPrintJsonLines(respData))
	} else {
		log.Printf("[ERROR] %s API Response error: %#v", t.name, err)
	}

	return resp, nil
}

func newDebugTransport(t http.RoundTripper) *debugTransport {
	return &debugTransport{name: "Algolia", transport: t}
}

// prettyPrintJsonLines iterates through a []byte line-by-line,
// transforming any lines that are complete json into pretty-printed json.
func prettyPrintJsonLines(b []byte) string {
	parts := strings.Split(string(b), "\n")
	for i, p := range parts {
		if b := []byte(p); json.Valid(b) {
			var out bytes.Buffer
			if err := json.Indent(&out, b, "", " "); err != nil {
				continue
			}
			parts[i] = out.String()
		}
		// Mask following header values
		// X-Algolia-Api-Key
		// X-Algolia-Application-Id
		if strings.Contains(strings.ToLower(p), "x-algolia") {
			kv := strings.Split(p, ": ")
			if len(kv) != 2 {
				continue
			}
			kv[1] = strings.Repeat("*", len(kv[1]))
			parts[i] = strings.Join(kv, ": ")
		}
	}
	return strings.Join(parts, "\n")
}

const logReqMsg = `%s API Request Details:
---[ REQUEST ]---------------------------------------
%s
-----------------------------------------------------`

const logRespMsg = `%s API Response Details:
---[ RESPONSE ]--------------------------------------
%s
-----------------------------------------------------`
