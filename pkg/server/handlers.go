package server

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

func (s *Server) initTemplates() {
	tmpl, err := template.New("429").Parse(" <html>\n<head>\n<title>Too Many Requests</title>\n</head>\n<body>\n<h1>Too Many Requests</h1>\n" +
		"<p>It's only allowed {{ .RequestLimit}} requests per {{ .Minutes}} to this Web site per\nsubnet.  Try again soon.</p>\n      </body>\n   </html>")
	if err != nil {
		log.Fatal(err)
	}
	s.toManyReqTempl = *tmpl
}

func (s *Server) resetHandler(writer http.ResponseWriter, request *http.Request) {

	header, ok := request.Header["X-Forwarded-For"]
	if !ok || len(header) == 0 {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("bad request : empty X-Forwarded-For header"))
		return
	}

	err := s.service.ResetPrefixForIpv4(header[0])
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (s *Server) mainHandler(fs http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		header, ok := request.Header["X-Forwarded-For"]
		if !ok || len(header) == 0 {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("bad request : empty X-Forwarded-For header"))
			return
		}

		isBlocked, err := s.service.IsLimitExceededForIp(header[0])
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}

		if isBlocked {
			writer.WriteHeader(http.StatusTooManyRequests)
			err = s.toManyReqTempl.Execute(writer, struct {
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
