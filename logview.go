package embedweb

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Handler log view 入口
func LobViewHandler(title string, logRoot string) http.HandlerFunc {
	var logHandler = map[string]http.HandlerFunc{
		"menu": handleLogMenu(logRoot),
		"view": handleLogView(),
	}

	return func(w http.ResponseWriter, req *http.Request) {
		info := req.FormValue("info")
		host := req.FormValue("host")

		if info == "" {
			toHtml(w, `
				<html xmlns="http://www.w3.org/1999/xhtml">
					<head>
						<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
						<title>{{.title}}</title>
					</head>
					<frameset cols="300,*">
						<frame src="{{.routePath}}?info=menu&host={{.host}}" name="menu">
						<frame src="{{.routePath}}?info=result&host={{.host}}" name="result">
					</frameset>
				</html>`, X{
				"routePath": req.URL.Path,
				"host":      host,
				"title":     title,
			})
			return
		}

		handler, found := logHandler[req.FormValue("info")]
		if !found {
			handler = func(w http.ResponseWriter, req *http.Request) {
				fmt.Fprint(w, "unknow info")
			}
			return
		}

		handler(w, req)
	}
}

func handleLogMenu(logRoot string) http.HandlerFunc {
	if logRoot == "" {
		logRoot = "/tmp"
	}

	return func(w http.ResponseWriter, req *http.Request) {
		type Dirs struct {
			Name     string
			Pathname string
		}

		pathname := req.FormValue("pathname")
		if pathname == "" {
			if runtime.GOOS == "windows" {
				p, _ := os.Getwd()
				pathname = p
			} else {
				pathname = logRoot
			}
		}
		data := X{"routePath": req.URL.Path, "pathname": pathname, "host": req.FormValue("host")}

		if f, err := os.Open(pathname); err != nil {
			data["error"] = err.Error()
		} else {
			if info, err := f.Stat(); err != nil {
				f.Close()
				data["error"] = err.Error()
			} else {
				if info.IsDir() {
					if infos, err := f.Readdir(0); err != nil {
						data["error"] = err.Error()
					} else {
						dirs := make([]Dirs, 0, len(infos))
						logs := make([]Dirs, 0, len(infos))
						for _, info := range infos {
							if info.IsDir() {
								dirs = append(dirs, Dirs{info.Name(), filepath.Join(pathname, info.Name())})
							} else {
								logs = append(logs, Dirs{info.Name(), filepath.Join(pathname, info.Name())})
							}
						}
						data["dirs"] = dirs
						data["logs"] = logs
					}
					f.Close()
				}
			}
		}

		if err := toHtml(w, `
			<h4>Index of: {{.pathname}}</h4>
			<hr />
			<h3>{{.error}}</h3>
			<ul style="list-style-type:circle;">
				{{range .dirs}}
		        <li><a href="{{$.routePath}}?info=menu&host={{$.host}}&pathname={{.Pathname}}" target="menu">{{.Name}}</a></li>
		        {{end}}
			</ul>
			{{if gt (len .dirs) 0}}<hr />{{end}}
			<ul>
				{{range .logs}}
		        <li><a href="{{$.routePath}}?info=view&host={{$.host}}&pathname={{.Pathname}}" target="result">{{.Name}}</a></li>
		        {{end}}
			</ul>
			<pre>{{.content}}<pre>`, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func handleLogView() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const pageSize = 4096
		var (
			pathname = req.FormValue("pathname")
			logName  = pathname
			pos, _   = strconv.ParseInt(req.FormValue("pos"), 10, 0)
			buf      = make([]byte, pageSize, pageSize)
			data     = X{
				"routePath":   req.URL.Path,
				"pathname":    pathname,
				"host":        req.FormValue("host"),
				"fileSize":    0,
				"curPos":      pos,
				"prePagePos":  0,
				"nextPagePos": 0,
				"page":        1,
				"pages":       0,
				"content":     "",
			}
		)

		if i := strings.LastIndex(pathname, string(os.PathSeparator)); i != -1 {
			logName = pathname[i+1:]
		}
		data["logname"] = logName

		defer func() {

			if err := toHtml(w, `
				{{define "nav"}}<pre>{{if gt .curPos 0}}<a href="{{.routePath}}?info=view&host={{.host}}&pos=0&pathname={{.pathname}}">首页</a> | `+
				`<a href="{{.routePath}}?info=view&pos={{.prePagePos}}&host={{.host}}&pathname={{.pathname}}">上一页</a> | {{end}}`+
				`{{if lt .nextPagePos .fileSize}}<a href="{{.routePath}}?info=view&host={{.host}}&pos={{.nextPagePos}}&pathname={{.pathname}}">下一页</a> |{{end}} `+
				`<a href="{{.routePath}}?info=view&pos=-1&host={{.host}}&pathname={{.pathname}}">最后一页</a> [{{.page}}/{{.pages}}]</pre>{{end}}

				<h3>{{.logname}}</h3>
				<h3>{{.error}}</h3>{{template "nav" .}}
				<hr />
 	            <pre>{{.content}}</pre>
 	            <hr />
 	            {{template "nav" .}}`,
				data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()

		f, err := os.Open(pathname)
		if err != nil {
			data["error"] = err
			return
		}
		defer f.Close()

		fileInfo, err := f.Stat()
		if err != nil {
			data["error"] = err
			return
		}

		if pos < 0 {
			if fileInfo.Size() < pageSize {
				pos = 0
			} else {
				pos = fileInfo.Size() - pageSize
			}
		} else if pos >= fileInfo.Size() {
			pos = fileInfo.Size()
		}
		f.Seek(pos, os.SEEK_SET)

		n, err := f.Read(buf)
		if err != nil {
			data["error"] = err
			return
		}

		nextPagePos := pos + int64(n)
		if nextPagePos > fileInfo.Size() {
			nextPagePos = fileInfo.Size()
		}
		prePagePos := pos - pageSize
		if prePagePos < 0 {
			prePagePos = 0
		}

		content := string(buf[:n])

		data["fileSize"] = fileInfo.Size()
		data["curPos"] = pos
		data["prePagePos"] = prePagePos
		data["nextPagePos"] = nextPagePos
		data["page"] = pos/pageSize + 1
		data["pages"] = fileInfo.Size()/pageSize + func() int64 {
			if (fileInfo.Size() % pageSize) > 0 {
				return 1
			}
			return 0
		}()
		data["content"] = content
	}
}
