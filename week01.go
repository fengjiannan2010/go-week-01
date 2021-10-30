package main

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"go-week-01/model"
	"log"
	"net/http"
	"os"
	"time"
)

type DbConf struct {
	USERNAME string `json:"username,default=root"`
	PASSWORD string `json:"password,default=root"`
	HOST     string `json:"host,default=localhost"`
	PORT     string `json:"port,default=3260"`
	DATABASE string `json:"database,default=test01"`
	CHARSET  string `json:"charset"`
}

var MysqlDb *sql.DB

func main() {
	dbconf := DbConf{
		USERNAME: "root",
		PASSWORD: "root",
		HOST:     "localhost",
		PORT:     "3260",
		DATABASE: "test01",
		CHARSET:  "utf-8",
	}
	sqldb, err := InitMysql(dbconf)
	if err != nil {
		log.Printf("Init Mysql Error:%s", err.Error())
		return
	}
	MysqlDb = sqldb
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/user/login", userLoginHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Printf("Open http://localhost:%s in the browser", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_, err := fmt.Fprint(w, "Hello, World!")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func userLoginHandler(w http.ResponseWriter, r *http.Request) {
	userName := r.FormValue("userName")
	_, err := login(userName)
	if err != nil {
		_, err := fmt.Fprint(w, fmt.Sprintf("%+v", errors.Cause(err)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		_, err := fmt.Fprint(w, "Login Success")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func login(userName string) (bool, error) {
	var user model.User
	// 查询 QueryRow 返回一条
	row := MysqlDb.QueryRow("SELECT * FROM  user_info WHERE `user_name` = ?", userName)
	err := row.Scan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			err := errors.Wrap(err, "登录失败，用户不存在")
			return false, err
		}
		return false, err
	}
	return true, nil
}

func InitMysql(conf DbConf) (*sql.DB, error) {
	dbDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", conf.USERNAME, conf.PASSWORD, conf.HOST, conf.PORT, conf.DATABASE, conf.CHARSET)
	// 打开连接失败
	mysqlDb, err := sql.Open("mysql", dbDsn)
	if err != nil {
		panic("数据源配置错误: " + err.Error())
	}
	// 最大连接数
	mysqlDb.SetMaxOpenConns(100)
	// 闲置连接数
	mysqlDb.SetMaxIdleConns(20)
	// 最大连接周期
	mysqlDb.SetConnMaxLifetime(100 * time.Second)
	if err = mysqlDb.Ping(); nil != err {
		panic("数据库链接失败: " + err.Error())
	}
	return mysqlDb, err
}
