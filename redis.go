package main

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/mattheath/base62"
	"log"
	"time"
)

const   (
	URLIDKEY = "next.url.id"
	ShortlinkKey = "shortlink:%s:url"
	URLHashKey = "urlhash:%s:url"
	ShortlinkDetailKey = "shortlink:%s:detail"
)

//RedisCli
type RedisCli struct {
	Cli *redis.Client
}

//UrlDetail:the details of the shortlink
type UrlDetail struct {
	URL string `json:"url"`
	CreatedAt string `json:"created_at"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

//NewRedisCli create a new redisclient
func NewRedisCli(add string,passwd string,db int) *RedisCli  {
	c := redis.NewClient(&redis.Options{
		Addr:add,
		Password:passwd,
		DB:db,
	})

	if _,err := c.Ping().Result();err != nil{
		log.Print(err)
		panic(err)
	}

	return &RedisCli{Cli:c}
}

//shorten convert url to shortlink
func (r *RedisCli) Shorten(url string,exp int64) (string,error) {
	h := toSha1(url)

	d,err := r.Cli.Get(fmt.Sprintf(URLHashKey,h)).Result()
	if err == redis.Nil{

	} else if err != nil{
		return "",err
	} else {
		if d == "{}"{

		} else {
			return d,nil
		}
	}

	err = r.Cli.Incr(URLIDKEY).Err()
	if err != nil{
		return "",err
	}

	//shortlink --> url
	id,err := r.Cli.Get(URLIDKEY).Int64()
	if err != nil{
		return "",err
	}
	eid := base62.EncodeInt64(id)
	log.Print(eid)

	err = r.Cli.Set(fmt.Sprintf(ShortlinkKey,eid),url,time.Minute*time.Duration(exp)).Err()
	if err != nil{
		return "",err
	}

	//urlhash --> url
	err = r.Cli.Set(fmt.Sprintf(URLHashKey,h),eid,time.Minute*time.Duration(exp)).Err()
	if err != nil{
		return "",err
	}
	
	detail,err := json.Marshal(
		&UrlDetail{
			URL:                 url,
			CreatedAt:           time.Now().String(),
			ExpirationInMinutes: time.Duration(exp),
		},
		)
	if err != nil{
		return "",err
	}
	//这里的detail是数字流[]byte格式
	log.Printf("%s",detail)

	err = r.Cli.Set(fmt.Sprintf(ShortlinkDetailKey,eid),detail,time.Minute*time.Duration(exp)).Err()
	if err != nil{
		return "",err
	}

	return eid,err
}

func toSha1(str string) string {
	sha := sha1.New()
	return string(sha.Sum([]byte(str)))
}

//Shortlinkinfo
func (r *RedisCli) ShortlinkInfo(eid string) (interface{},error) {
	d,err := r.Cli.Get(fmt.Sprintf(ShortlinkDetailKey,eid)).Result()
	if err == redis.Nil{
		return "",StatusError{
			Code: 404,
			Err:  errors.New("Unknown short url"),
		}
	} else if err != nil{
		return "",err
	}
	return d,nil
}

//Unshorten
func (r *RedisCli) Unshorten(eid string)(string,error) {
	url ,err := r.Cli.Get(fmt.Sprintf(ShortlinkKey,eid)).Result()
	if err == redis.Nil{
		return "",StatusError{
			Code: 404,
			Err:  errors.New("Incorrect short url"),
		}
	} else if err != nil{
		return "",err
	}

	return url,nil

}