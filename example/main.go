package main

import (
	"net/http"
	"github.com/winxxp/embedweb"
)

func main()  {
	http.Handle("/app/proxy/log", embedweb.LobViewHandler("App logview", "/"))
	http.Handle("/app/proxy/shell", embedweb.ShellHandler("App shell"))
	http.Handle("/app/proxy/editor", embedweb.EditHandler("App editor"))
	
	http.ListenAndServe(":8080", nil)
}