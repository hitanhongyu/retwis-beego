package routers

import (
	web "github.com/astaxie/beego"
	"retwis/controllers"
)

func init() {
	web.Get("/",controllers.Index)
	web.Get("/home",controllers.Home)
	web.Post("/register", controllers.Register)
	web.Post("/login", controllers.Login)
	web.Get("/logout", controllers.Logout)
	web.Post("/post", controllers.Publish)
	web.Get("/timeline", controllers.Timeline)
	web.Get("/profile", controllers.Profile)
	web.Get("/follow", controllers.Follow)
}
