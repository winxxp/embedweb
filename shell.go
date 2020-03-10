package embedweb

import (
	"bytes"
	"context"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
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
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
			defer cancel()

			args := strings.Fields("/C " + cmd)
			buf, err := exec.CommandContext(ctx, "cmd", args...).CombinedOutput()

			if err != nil {
				data["result"] = err.Error() + GBK2UTF8(buf)
			} else {
				data["result"] = GBK2UTF8(buf)
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

func GBK2UTF8(gbk []byte) string {
	out := bytes.Buffer{}
	enc := transform.NewReader(bytes.NewBuffer(gbk), simplifiedchinese.GB18030.NewDecoder())
	if _, err := io.Copy(&out, enc); err != nil {
		return "N/A"
	}
	return out.String()
}
