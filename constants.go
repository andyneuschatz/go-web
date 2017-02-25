package web

const (
	HeaderAcceptEncoding = "Accept-Encoding"

	HeaderDate         = "Date"
	HeaderCacheControl = "Cache-Control"

	HeaderConnection = "Connection"

	HeaderContentEncoding = "Content-Encoding"
	HeaderContentLength   = "Content-Length"
	HeaderContentType     = "Content-Type"
	HeaderContentEncoding = "Content-Encoding"

	HeaderServer = "Server"
	HeaderVary   = "Vary"

	HeaderXServedBy           = "X-Served-By"
	HeaderXFrameOptions       = "X-Frame-Options"
	HeaderXXSSProtection      = "X-Xss-Protection"
	HeaderXContentTypeOptions = "X-Content-Type-Options"

	// ContentTypeApplicationJSON is a content type for JSON responses.
	ContentTypeApplicationJSON = "application/json; charset=UTF-8"

	// ContentTypeHTML is a content type for html responses.
	ContentTypeHTML = "text/html; charset=utf-8"

	//ContentTypeXML is a content type for XML responses.
	ContentTypeXML = "text/xml; charset=utf-8"

	// ContentTypeText is a content type for text responses.
	ContentTypeText = "text/plain; charset=utf-8"

	// ConnectionKeepAlive is a value for the "Connection" header and
	// indicates the server should keep the tcp connection open
	// after the last byte of the response is sent.
	ConnectionKeepAlive = "keep-alive"
)
