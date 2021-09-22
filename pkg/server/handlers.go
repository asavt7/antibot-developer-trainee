package server

import (
	"errors"
	"html/template"
	"log"
	"net"
	"net/http"
	"time"
)

var ToManyReqTemplate *template.Template

func init() {
	initTemplates()
}

func initTemplates() {
	tmpl, err := template.New("429").Parse(" <html>\n<head>\n<title>Too Many Requests</title>\n</head>\n<body>\n<h1>Too Many Requests</h1>\n" +
		"<p>It's only allowed {{ .RequestLimit}} requests per {{ .Minutes}} to this Web site per\nsubnet.  Try again soon.</p>\n      </body>\n   </html>")
	if err != nil {
		log.Fatal(err)
	}
	ToManyReqTemplate = tmpl
}

func (s *Server) resetHandler(writer http.ResponseWriter, request *http.Request) {
	ipv4, err := parseHeaderXForwardedFor(request.Header)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}

	err = s.service.ResetPrefixForIpv4(ipv4)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func parseHeaderXForwardedFor(headers http.Header) (net.IP, error) {
	header, ok := headers["X-Forwarded-For"]
	if !ok || len(header) == 0 {
		return nil, errors.New("bad request : empty X-Forwarded-For header")
	}
	ipv4 := net.ParseIP(header[0]).To4()
	if ipv4 == nil {
		return nil, errors.New("bad request : invalid X-Forwarded-For header value - expected IPv4 address")
	}
	return ipv4, nil
}

func (s *Server) mainHandler(fs http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.RequestURI != "/" {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		ipv4, err := parseHeaderXForwardedFor(request.Header)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(err.Error()))
			return
		}

		isBlocked, err := s.service.IsLimitExceededForIp(ipv4)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}

		if isBlocked {
			writer.WriteHeader(http.StatusTooManyRequests)
			err = ToManyReqTemplate.Execute(writer, struct {
				RequestLimit int
				Minutes      time.Duration
			}{
				RequestLimit: s.config.RequestLimit,
				Minutes:      s.config.TimeInterval,
			})
			if err != nil {
				log.Println(err.Error())
			}
			return
		}
		fs.ServeHTTP(writer, request)
	}
}
