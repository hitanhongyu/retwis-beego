package models

import (
	"errors"
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/securecookie"
	"github.com/kylemcc/twitter-text-go/extract"
	"strconv"
	"strings"
	"time"
)

// IsLogin return the user who have auth is login or not
func IsLogin(auth string)(*User ,error){
	if len(auth)==0{
		return nil,errors.New("No authentification token")
	}
	userId,err := redis.String(Conn.Do("HGET","auths",auth))
	if err!=nil{
		return nil, err
	}
	saveAuth,err := redis.String(Conn.Do("HGET","user:"+userId,auth))
	if err!=nil || saveAuth!=auth{
		return nil, errors.New("Wrong authentification token")
	}
	return nil, nil
}
//GetUsrInfo return the info of usr by its userId
func GetUsrInfo(userId string)(*User ,error){
	if userId==""{
		return nil, errors.New("wrong userId")
	}
	v,err:=redis.Values(Conn.Do("HGETALL","user:"+userId))
	if err!=nil{
		return nil, err
	}
	u:=&User{Id:userId}
	err = redis.ScanStruct(v,u)
	if err!=nil{
		return nil, err
	}
	return u,nil
}

//profileByUsername return user by its name
func ProfileByUsername(username string) (*User, error){
	if username==""{
		return nil,errors.New("Wrong username")
	}
	userId,err:=redis.String(Conn.Do("HGET","users",username))
	if err!=nil{
		return nil, err
	}
	u:=&User{
		Id: userId,
		Username: username,
	}
	return u, nil
}

//profileByUserId return user by its userId
func ProfileByUserId(userId string) (*User, error) {

	if userId == "" {
		return nil, errors.New("Invalid user Id")
	}
	username, err := redis.String(Conn.Do("HGET", "user:"+userId, "username"))
	if err != nil {
		return nil, err
	}
	u := &User{Id: userId, Username: username}
	return u, nil
}


//register return the auth in the register period
func Register(username, password string) (auth string, err error) {

	userId, err := redis.String(Conn.Do("INCR", "next_user_id"))
	if err != nil {
		return "", err
	}
	/*
		todo
	*/
	exist,err:=redis.Int(Conn.Do("EXISTS","user:"+userId))
	if exist==1{
		return "",errors.New("user already exists")
	}
	auth = string(securecookie.GenerateRandomKey(32))
	Conn.Do("HSET","users",username,userId)
	Conn.Do("HMSET","user:"+userId,"username",username,"password",password,"auth",auth)
	Conn.Do("HSET","auths",auth,userId)
	Conn.Do("ZADD","users_by_time",time.Now().Unix(),username)

	return auth,nil
}
//login return the auth
func Login(username string,passwd string) (auth string,err error){
	userId,err:=redis.String(Conn.Do("HGET","users",username))
	if err!=nil{
		return "", err
	}
	realPasswd,err:=redis.String(Conn.Do("HGET","user:"+userId,"password"))
	if err!=nil{
		return "", err
	}
	if realPasswd!=passwd{
		return "",errors.New("Wrong Password")
	}
	auth,err = redis.String(Conn.Do("HGET","user:"+userId,"auth"))
	if err!=nil{
		return "", nil
	}
	return auth,nil
}

//logout
func Logout(user *User){
	if user==nil{
		return
	}
	userId := user.Id
	newAuth := string(securecookie.GenerateRandomKey(32))
	oldAuth,_ := redis.String(Conn.Do("HGET","user:"+userId,"auth"))

	_,err:=Conn.Do("HSET","user:"+userId,"auth",newAuth)
	_,err = Conn.Do("HSET","auths",newAuth,userId)
	_,err = Conn.Do("HDEL","auths",oldAuth)
	if err!=nil {
		return
	}
}
//Post
func To_Post(user *User,status string) error{
	postId,err:=redis.Int(Conn.Do("INCR","next_post_id"))
	if err!=nil{
		return err
	}
	status = strings.Replace(status,"\n"," ",-1)
	_,err = Conn.Do("HMSET",fmt.Sprintf("post:%d",postId),"user_id",user.Id,"time",time.Now().Unix(),"body",status)
	if err!=nil{
		return err
	}
	//add the usr who is following
	followers,err:=redis.Values(Conn.Do("ZRANGE","followers:"+user.Id,0.-1))
	if err!=nil{
		return err
	}
	recipients := mapset.NewSet()
	for _,fId:=range followers{
		recipients.Add(fId)
	}
	//add the @one
	entities := extract.ExtractMentionedScreenNames(status)
	for _,e:=range entities{
		//todo find out what in the entitles
		username,_:= e.ScreenName()
		profile,err:=ProfileByUsername(username)
		if err==nil{
			recipients.Add(profile.Id)
		}
 	}
 	//add itself
 	recipients.Add(user.Id)
    for _,rId:=range recipients.ToSlice(){
    	str,ok:=rId.(string)
    	if ok{
    		Conn.Do("LPUSH","posts:"+str,postId)
		}
	}
	_, err = Conn.Do("LPUSH","timeline",postId)
	if err!=nil{
		return err
	}
	// timeline limits 1000
	_, err = Conn.Do("LTRIM", "timeline", 0, 1000)
	if err != nil {
		return err
	}
	return nil
}

func strElapsed(t string) string {

	ts, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return ""
	}
	te := time.Now().Unix() - ts
	if te < 60 {
		return fmt.Sprintf("%d seconds", te)
	}
	if te < 3600 {
		m := int(te / 60)
		if m > 1 {
			return fmt.Sprintf("%d minutes", m)
		} else {
			return fmt.Sprintf("%d minute", m)
		}
	}
	if te < 3600*24 {
		h := int(te / 3600)
		if h > 1 {
			return fmt.Sprintf("%d hours", h)
		} else {
			return fmt.Sprintf("%d hour", h)
		}
	}
	d := int(te / (3600 * 24))
	if d > 1 {
		return fmt.Sprintf("%d days", d)
	} else {
		return fmt.Sprintf("%d day", d)
	}
}
//getPost  return the Post by its id
func getPost(postId string) (*Post ,error){
	v,err:=redis.Values(Conn.Do("HGETALL","post:"+postId))
	if err!=nil{
		return nil, err
	}
	p:=&Post{}
	err = redis.ScanStruct(v,p)
	if err!=nil{
		return nil, err
	}
	username,err:=redis.String(Conn.Do("HGET","user:"+p.UserID,"username"))
	p.Username=username
	p.Elapsed=strElapsed(p.Elapsed)
	return p,nil
}
/*
	key :post+postId
*/
func GetUserPosts(key string ,start, count int64) ([]*Post, int64, error){

	values,err := redis.Strings(Conn.Do("LRANGE",key,start,start+count-1))
	if err!=nil{
		return nil,0,err
	}
 	posts := make([]*Post,0)
 	for _,pid:= range values{
 		p ,err:=getPost(pid)
 		if err!=nil{
 			posts = append(posts,p)
		}
	}
	l,err := redis.Int64(Conn.Do("LLEN",key))
	if err !=nil{
		return nil,0,err
	}else{
		return posts,l - start -int64(len(values)),nil
	}
}

func GetLastUsers() ([]*User,error){
	//get the last 10 user
	v, err := redis.Strings(Conn.Do("ZREVRANGE", "users_by_time", 0, 9))
	if err!=nil{
		return nil, err
	}
	users:=make([]*User,0)
	for _,username:=range v{
		users = append(users,&User{
			Username: username,
		})
	}
	return users, nil
}