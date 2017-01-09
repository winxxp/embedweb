package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := r.FormValue("host")
		if host == "" {
			http.Error(w, "host must be input", http.StatusBadRequest)
			return
		}

		targetUrl := url.URL{
			Scheme:   "http",
			Host:     host,
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		}

		if r.Method == "GET" {
			resp, err := http.Get(targetUrl.String())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.Copy(w, resp.Body)
		} else if r.Method == "POST" {
			body := r.Form.Encode()
			resp, err := http.Post(targetUrl.String(), r.Header.Get("Content-Type"), bytes.NewBufferString(body))
			if err != nil {
				http.Error(w, "unsupport method:"+r.Method, http.StatusMethodNotAllowed)
				return
			}
			io.Copy(w, resp.Body)
		} else {
			http.Error(w, "unsupport method:"+r.Method, http.StatusMethodNotAllowed)
		}
	}
}
