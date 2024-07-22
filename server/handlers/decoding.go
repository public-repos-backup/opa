package handlers

import (
	"net/http"
	"strings"

	"github.com/open-policy-agent/opa/server/types"
	"github.com/open-policy-agent/opa/server/writer"
	util_decoding "github.com/open-policy-agent/opa/util/decoding"
)

// This handler provides hard limits on the size of the request body, for both
// the raw body content, and also for the decompressed size when gzip
// compression is used.
//
// The Content-Length restriction happens here in the handler, but the
// decompressed size limit is enforced later, in `util.ReadMaybeCompressedBody`.
// The handler passes the gzip size limits down to that function through the
// request context whenever gzip encoding is present.
func DecodingLimitsHandler(handler http.Handler, maxLength, gzipMaxLength int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reject too-large requests before doing any further processing.
		// Note(philipc): This likely does nothing in the case of "chunked"
		// requests, since those should report a ContentLength of -1.
		if r.ContentLength > maxLength {
			writer.Error(w, http.StatusBadRequest, types.NewErrorV1(types.CodeInvalidParameter, types.MsgDecodingLimitError))
			return
		}
		// Pass server.decoding.gzip.max_length down, using the request context.
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			ctx := util_decoding.AddServerDecodingGzipMaxLen(r.Context(), gzipMaxLength)
			r = r.WithContext(ctx)
		}

		// Copied over from the net/http package; enforces max body read limits.
		r2 := *r
		r2.Body = http.MaxBytesReader(w, r.Body, maxLength)
		handler.ServeHTTP(w, &r2)
	})
}
