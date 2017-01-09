package embedweb

import (
	"net/http"
	"os/exec"
	"strings"
)

// Handler log view 入口
func ShellHandler(title string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cmd := req.FormValue("cmd")
		data := X{
			"host": req.FormValue("host"),
			"cmd":  cmd,
		}

		if cmd != "" {
			args := strings.Fields(cmd)
			buf, err := exec.Command(args[0], args[1:]...).CombinedOutput()
			if err != nil {
				data["result"] = err.Error() + string(buf)
			} else {
				data["result"] = string(buf)
			}
		}

		if err := toHtml(w, `
				<html xmlns="http://www.w3.org/1999/xhtml">
					<head>
						<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
						<title>{{.title}}</title>
					</head>
					<body>
						<h2>执行系统shell，请小心使用!</h2>
						<form action="{{.routepath}}" method="post" enctype="application/x-www-form-urlencoded">
							<input name="cmd" type="text" value="" maxlength="255" style="width:89%" />
							<input name="host" type="hidden" value="{{.host}}" />
							<input type="submit" value="执行" style="width:10%" />
						</form>
						<h4># {{.cmd}}<br></h4>
						<pre>{{.result}}</pre>
					</body>
				</html>
			`, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
