package controllers

import (
	"errors"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/unrolled/render"
	"log"
	"net/http"
	"retwis/cookie"
	"retwis/models"
	"strconv"
)

type MainController struct {
	beego.Controller
}

func (this *MainController)Get(){
	this.TplName="welcome"
	ctx:=this.Ctx
	templateParams := map[string]interface{}{}
	u, err := models.IsLogin(cookie.GetAuth(ctx.Request))
	templateParams["user"] = u
	if err!=nil{
		 this.Render()
	}else{
		this.Redirect("/home",http.StatusFound)
	}
}
var (
	templateRender = render.New(render.Options{
		Layout:        "layout",
		IsDevelopment: true,
	})
)

func Index(ctx *context.Context) {
	templateParams := map[string]interface{}{}
	u, err := models.IsLogin(cookie.GetAuth(ctx.Request))
	templateParams["user"] = u
	if err != nil {

	} else {
		ctx.Redirect(http.StatusFound, "/home")
	}
}

func Home(ctx *context.Context) {
	templateParams := map[string]interface{}{}
	u, err := models.IsLogin(cookie.GetAuth(ctx.Request))
	if nil != err {
		http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
		return
	}
	templateParams["user"] = u

	var start int64
	if "" == ctx.Request.FormValue("start") {
		start = int64(0)
	} else {
		start, err = strconv.ParseInt(ctx.Request.FormValue("start"), 10, 64)
		if err != nil {
			start = int64(0)
		}
	}
	posts, rest, err := models.GetUserPosts("posts:"+u.Id, start, 10)
	if err == nil {
		if start > 0 {
			templateParams["prev"] = start - 10
		}
		templateParams["posts"] = posts
		if rest > 0 {
			templateParams["next"] = start + 10
		}
	}
	templateRender.HTML(ctx.ResponseWriter, http.StatusOK, "home", templateParams)
}

func Register(ctx *context.Context) {
	username := ctx.Request.PostFormValue("username")
	password := ctx.Request.PostFormValue("password")
	password2 := ctx.Request.PostFormValue("password2")
	if username == "" || password == "" || password2 == "" {
		GoBack(ctx.ResponseWriter, ctx.Request, errors.New("Every field of the registration form is needed!"))
		return
	}
	if password != password2 {
		GoBack(ctx.ResponseWriter, ctx.Request, errors.New("The two password fileds don't match!"))
		return
	}
	auth, err := models.Register(username, password)
	if err != nil {
		GoBack(ctx.ResponseWriter, ctx.Request, err)
		return
	}
	cookie.SetSession(auth, ctx.ResponseWriter)
	templateParams := map[string]interface{}{}
	templateParams["username"] = username
	templateRender.HTML(ctx.ResponseWriter, http.StatusOK, "register", templateParams)
}

func Login(ctx *context.Context) {
	username := ctx.Request.PostFormValue("username")
	password := ctx.Request.PostFormValue("password")
	if username == "" || password == "" {
		GoBack(ctx.ResponseWriter, ctx.Request, errors.New("You need to enter both username and password to login."))
		return
	}
	auth, err := models.Login(username, password)
	if err != nil {
		GoBack(ctx.ResponseWriter, ctx.Request, err)
		return
	}
	cookie.SetSession(auth, ctx.ResponseWriter)
	http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
}

func Logout(ctx *context.Context) {
	u, err := models.IsLogin(cookie.GetAuth(ctx.Request))
	if nil != err {
		http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
		return
	}
	models.Logout(u)
	http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
}

func Publish(ctx *context.Context) {
	u, err := models.IsLogin(cookie.GetAuth(ctx.Request))
	if nil != err {
		http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
		return
	}
	status := ctx.Request.PostFormValue("status")
	if status == "" {
		http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
		return
	}
	err = models.To_Post(u, status)
	http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
}

func Timeline(ctx *context.Context) {
	templateParams := map[string]interface{}{}
	users, err := models.GetLastUsers()
	if err != nil {
		log.Println(err)
	} else {
		templateParams["users"] = users
	}
	posts, _, err := models.GetUserPosts("timeline", 0, 50)
	if err != nil {
		log.Println(err)
	} else {
		templateParams["posts"] = posts
	}
	templateRender.HTML(ctx.ResponseWriter, http.StatusOK, "timeline", templateParams)
}

func Profile(ctx *context.Context) {
	templateParams := map[string]interface{}{}
	// get username
	username := ctx.Request.FormValue("u")
	if username == "" {
		http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
		return
	}
	// Get profile
	p, err := models.ProfileByUsername(username)
	if err != nil {
		http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
		return
	}
	templateParams["profile"] = p
	// Get logged in user
	u, err := models.IsLogin(cookie.GetAuth(ctx.Request))
	if nil == err {
		templateParams["user"] = u
	}

	var start int64
	if "" == ctx.Request.FormValue("start") {
		start = int64(0)
	} else {
		start, err = strconv.ParseInt(ctx.Request.FormValue("start"), 10, 64)
		if err != nil {
			start = int64(0)
		}
	}
	posts, rest, err := models.GetUserPosts("posts:"+p.Id, start, 10)
	if err == nil {
		if start > 0 {
			templateParams["prev"] = start - 10
		}
		templateParams["posts"] = posts
		if rest > 0 {
			templateParams["next"] = start + 10
		}
	}
	templateRender.HTML(ctx.ResponseWriter, http.StatusOK, "profile", templateParams)
}

func Follow(ctx *context.Context) {
	// get the user id
	userId := ctx.Request.FormValue("uid")
	if userId == "" {
		http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
		return
	}
	// Get the action to do
	doFollow := false
	switch ctx.Request.FormValue("f") {
	case "1":
		doFollow = true
	case "0":
		doFollow = false
	default:
		http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
		return
	}
	u, err := models.IsLogin(cookie.GetAuth(ctx.Request))
	if nil != err {
		http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
		return
	}
	if userId == u.Id {
		http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
		return
	}
	if doFollow {
		u.Follow(&models.User{Id: userId})
	} else {
		u.UnFollow(&models.User{Id: userId})
	}
	p, err := models.ProfileByUserId(userId)
	if err != nil {
		http.Redirect(ctx.ResponseWriter, ctx.Request, "/", http.StatusFound)
		return
	}
	http.Redirect(ctx.ResponseWriter, ctx.Request, "/profile?u="+p.Username, http.StatusFound)
}

func GoBack(w http.ResponseWriter, r *http.Request, err error) {
	templateParams := map[string]interface{}{}
	templateParams["error"] = err
	templateRender.HTML(w, http.StatusOK, "error", templateParams)
}

