/*
彩云影视线上服务器源代码
v2.8.0以后版本使用
基本文件 main.go update.go tvlives.go videofetch.go videoRecom.go safecheck.go
*/

package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func TvServer() {
	router := mux.NewRouter()
	// 验证请求
	router.Handle("/search", netMiddleWare(HandleSearchVideo())).Methods("GET")
	router.Handle("/videolink", netMiddleWare(HandleShowVideoPage())).Methods("GET")
	router.Handle("/tvstream", netMiddleWare(HandlerTVStreamRequest())).Methods("GET")
	router.Handle("/tvlabel", netMiddleWare(HandlerTVLabelsRequest())).Methods("GET")
	router.Handle("/checkCodes", netMiddleWare(HandlerTransformHashCode())).Methods("GET")
	router.Handle("/fetchCodes", netMiddleWare(HandlerTransCodeList())).Methods("GET")
	// 一般请求
	router.HandleFunc("/recomVideo", HandlerRecomNewVideos()).Methods("GET")
	router.Handle("/infos", HandlerVideoInfos()).Methods("GET")
	router.Handle("/update", HanlderUpdateRequest()).Methods("GET")
	router.Handle("/mobile", HandlerUpdateMobile()).Methods("GET")
	err := http.ListenAndServe(":7215", router)
	if err != nil {
		log.Println(err.Error())
	}
}

func main() {
	initSafeCode()
	initSecretCode()
	initTVInfos()
	TvServer()
}
