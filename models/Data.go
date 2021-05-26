package models

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

type User struct {
	Id string
	Username string `redis:"username"`
	Auth string `redis:"auth"`
}

type Post struct {
	UserID string `redis:"user_id"`
	Username string
	Body string `redis:"body"`
	Elapsed string  `redis:"time"`
}

//Is return compare two users to verify if they are the same one.
func (user *User)Is(aUser *User) bool{
	if user == nil|| aUser ==nil{
		return false
	}
	if user.Id == aUser.Id{
		return true
	}
	return false
}
const (
	address = "127.0.0.1:6379"
)
var Conn redis.Conn
func init(){
	var err error
	Conn,err = redis.Dial("tcp" ,address)

	if err!=nil{
		fmt.Println(err)
	}else{
		fmt.Println("redis conn successfully!")
	}
}
//IsFollowing : return user p is following user u or not.
func (u *User)IsFollowing(p *User) bool{
	v,err:=redis.Int(Conn.Do("ZSCORE","following:"+u.Id,p.Id))
	if err!=nil{
		return false
	}
	if v>0 {
		return true
	}else{
		return false
	}
}

//Followers return the num of users who follow user u
func (u *User)Followers() int{
	followers,err:=redis.Int(Conn.Do("ZCARD","followers:"+u.Id))
	if  err !=nil{
		return  0
	}else{
		return followers
	}
}
// Following return the num of users who are followed by usr u.
func (u *User) Following() int{
	tofollowers,err := redis.Int(Conn.Do("ZCARD","following"+u.Id))
	if err!=nil{
		return 0
	}else{
		return tofollowers
	}
}

//Follow return user u follow user p
func (u *User)Follow(p *User){
	Conn.Do("ZADD","followers:"+p.Id,time.Now().Unix(),u.Id)
	Conn.Do("ZADD","following:"+u.Id,time.Now().Unix(),p.Id)
}

//UnFollow return usr unfollow usr p
func (u *User)UnFollow(p *User){
	Conn.Do("ZREM","followers"+p.Id,u.Id)
	Conn.Do("ZREM","following"+u.Id,p.Id)
}
