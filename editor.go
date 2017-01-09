package embedweb

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func EditHandler(title string) http.HandlerFunc {
	var logHandler = map[string]http.HandlerFunc{
		"menu": handleEditMenu(),
		"edit": handleEdit(),
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

func handleEditMenu() http.HandlerFunc {
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
				pathname = "/"
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
		        <li><a href="{{$.routePath}}?info=edit&host={{$.host}}&pathname={{.Pathname}}" target="result">{{.Name}}</a></li>
		        {{end}}
			</ul>
			<pre>{{.content}}<pre>`, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func handleEdit() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		filename := req.FormValue("pathname")
		if filename == "" {
			http.Error(w, "no filename input", http.StatusBadRequest)
			return
		}

		op := req.FormValue("action")
		if op == "modify" {
			content := req.FormValue("filecontent")
			if err := ioutil.WriteFile(filename, []byte(content), os.ModePerm); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			toHtml(w, "修改成功!", X{})
		} else {
			toHtml(w, `
				<form method="post" action="{{.routePath}}">
					<input name="modify" type="submit" value="Modify"><br>
					<input name="pathname" type="hidden" value="{{.filename}}">
					<input name="action" type="hidden" value="modify">
					<input name="host" type="hidden" value="{{.host}}">
					<input name="info" type="hidden" value="edit">
					<textarea name="filecontent" rows="60" ,="" cols="90">{{.content}}</textarea>
				</form>`, X{
				"routePath": req.URL.Path,
				"filename":  filename,
				"host":      req.FormValue("host"),
				"content": func() string {
					if buf, err := ioutil.ReadFile(filename); err != nil {
						return ""
					} else {
						return string(buf)
					}
				}(),
			})
		}
	}
}
