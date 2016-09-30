package main

import (
	"html/template"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	logging "github.com/op/go-logging"

	_ "github.com/wutongtree/funds/client/bootstrap"
	_ "github.com/wutongtree/funds/client/routers"
)

// var
var (
	logger = logging.MustGetLogger("funds.client")
)

func main() {
	beego.InsertFilter("/*", beego.BeforeRouter, filterUser)
	beego.ErrorHandler("404", pageNotFound)
	beego.ErrorHandler("401", pageNoPermission)
	beego.Run()
}

var filterUser = func(ctx *context.Context) {
	_, ok := ctx.Input.Session("userLogin").(string)
	if !ok && ctx.Request.RequestURI != "/login" {
		ctx.Redirect(302, "/login")
	}
}

func pageNotFound(rw http.ResponseWriter, r *http.Request) {
	t, _ := template.New("404.tpl").ParseFiles("views/404.tpl")
	data := make(map[string]interface{})
	t.Execute(rw, data)
}

func pageNoPermission(rw http.ResponseWriter, r *http.Request) {
	t, _ := template.New("401.tpl").ParseFiles("views/401.tpl")
	data := make(map[string]interface{})
	t.Execute(rw, data)
}
