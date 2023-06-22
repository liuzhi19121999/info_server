package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type TVInfosJson struct {
	Categories []string   `json:"cate"`
	Labels     [][]string `json:"labels"`
	Requests   [][]string `json:"reqs"`
}

type TVLiveStream struct {
	M3U8Stream map[string][]string `json:"stream"`
}

type TVInfoSum struct {
	TVLabel    TVInfosJson  `json:"tvlabel"`
	TVStream   TVLiveStream `json:"stream"`
	LabelTime  int64        `json:"labeltime"`
	StreamTime int64        `json:"streamtime"`
}

const timeTVLabelDuration = 1800
const timeTVStreamDuration = 3600

var tvInfos TVInfoSum

const tvinfosPath = "tvLabels.json"
const tvStreamPath = "tvStream.json"

func needUpdateInfos() {
	nowTime := time.Now().Unix()
	// 标签栏
	if (nowTime - tvInfos.LabelTime) >= timeTVLabelDuration {
		labelBytes, err := os.ReadFile(tvinfosPath)
		if err == nil {
			var tvlabs TVInfosJson
			err = json.Unmarshal(labelBytes, &tvlabs)
			if err == nil {
				tvInfos.TVLabel = tvlabs
				tvInfos.LabelTime = nowTime
			}
		}
	}
	if (nowTime - tvInfos.StreamTime) >= timeTVStreamDuration {
		m3u8Bytes, err := os.ReadFile(tvStreamPath)
		if err == nil {
			var tvstream TVLiveStream
			err = json.Unmarshal(m3u8Bytes, &tvstream)
			if err == nil {
				tvInfos.TVStream = tvstream
				tvInfos.StreamTime = nowTime
			}
		}
	}
}

func initTVInfos() {
	tvlabels := TVInfosJson{
		Categories: make([]string, 0),
		Labels:     make([][]string, 0),
		Requests:   make([][]string, 0),
	}
	tvLives := TVLiveStream{
		M3U8Stream: make(map[string][]string),
	}
	tvlabelBytes, err := os.ReadFile(tvinfosPath)
	if err == nil {
		err = json.Unmarshal(tvlabelBytes, &tvlabels)
		if err != nil {
			log.Println(err.Error())
		}
	}
	tvliveBytes, err := os.ReadFile(tvStreamPath)
	if err == nil {
		err = json.Unmarshal(tvliveBytes, &tvLives)
		if err != nil {
			log.Println(err.Error())
		}
	}
	tempTime := time.Now().Unix()
	tvInfos.LabelTime = tempTime
	tvInfos.StreamTime = tempTime
	tvInfos.TVLabel = tvlabels
	tvInfos.TVStream = tvLives
}

func HandlerTVLabelsRequest() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			needUpdateInfos()
			jsonBytes, err := json.Marshal(tvInfos.TVLabel)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err.Error())
				return
			}
			if secretCodeList.UseEnigma {
				jsonString := string(jsonBytes)
				//fmt.Println(jsonString)
				res := encodeds(jsonString)
				//fmt.Println(res)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(res))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(jsonBytes)
		},
	)
}

type SteamList struct {
	Streams []string `json:"streams"`
	ReqId   int      `json:"reqid"`
}

func HandlerTVStreamRequest() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			needUpdateInfos()
			var streams SteamList
			req := r.URL.Query().Get("channel")
			ids := r.URL.Query().Get("id")
			reqid, err := strconv.Atoi(ids)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err.Error())
				return
			}
			value, ok := tvInfos.TVStream.M3U8Stream[req]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			streams.Streams = value
			streams.ReqId = reqid
			jsonBytes, err := json.Marshal(streams)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err.Error())
				return
			}
			if secretCodeList.UseEnigma {
				jsonString := string(jsonBytes)
				res := encodeds(jsonString)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(res))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(jsonBytes)
		},
	)
}
