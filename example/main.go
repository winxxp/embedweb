package main

import (
	"github.com/winxxp/embedweb"
	"net/http"
)

func main() {
	lghandle := embedweb.LobViewHandler("App logview", "", embedweb.SN("1234"))
	http.Handle("/app/proxy/log", lghandle)
	http.Handle("/app/proxy/shell", embedweb.ShellHandler("App shell"))
	http.Handle("/app/proxy/editor", embedweb.EditHandler("App editor"))

	http.ListenAndServe(":8080", nil)
}
