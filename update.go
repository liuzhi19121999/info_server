package main

// 更新文档  兼  搜索栏下方影视推荐

import (
	"net/http"
	"os"
)

const updatePath = "updateTV.json"
const updateMobilePath = "updateMobile.json"

func HandlerUpdateMobile() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			jsonBytes, err := os.ReadFile(updateMobilePath)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(jsonBytes)
		},
	)
}

func HanlderUpdateRequest() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			//var updateInfos UpdateInfos
			jsonBytes, err := os.ReadFile(updatePath)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(jsonBytes)
		},
	)
}

const recomPath = "videoinfos.json"

func HandlerVideoInfos() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			jsonBytes, err := os.ReadFile(recomPath)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(jsonBytes)
		},
	)
}
