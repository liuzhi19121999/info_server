package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const fetchRecomTimeOut = time.Second * 15

type DouBanFetchItem struct {
	Epis     string `json:"episodes_info"`
	Cover    string `json:"cover"`
	CoverX   int    `json:"cover_x"`
	CoverY   int    `json:"cover_y"`
	Id       string `json:"id"`
	IsNew    bool   `json:"is_new"`
	Playable bool   `json:"playable"`
	Rate     string `json:"rate"`
	Title    string `json:"title"`
	Url      string `json:"url"`
}

type DouBanFetch struct {
	Subjects []DouBanFetchItem `json:"subjects"`
}

func FetchDoubanMovie(typeSearch string, cate string) (DouBanFetch, error) {
	douban := DouBanFetch{
		Subjects: make([]DouBanFetchItem, 0),
	}
	// 标头
	duban_url := "https://movie.com/j/search_subjects?type=" + typeSearch + "&tag=" + cate + "&page_limit=200&page_start=0"
	client := http.Client{Timeout: time.Duration(fetchRecomTimeOut)}
	req, err := http.NewRequest(http.MethodGet, duban_url, nil)
	if err != nil {
		return douban, err
	}
	req.Header.Add("Host", "movie.com")
	req.Header.Add("Referer", "https://movie.com/")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Edg/109.0.1518.55")
	r, err := client.Do(req)
	if err != nil {
		return douban, err
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&douban)
	if err != nil {
		return douban, err
	}
	return douban, nil
}

type RecomItem struct {
	Title string `json:"title"`
	Rate  string `json:"rate"`
	Image string `json:"img"`
}

type RecomList struct {
	TotalList []RecomItem `json:"total"`
}

func containList(list []string, val string) bool {
	for i := 0; i < len(list); i++ {
		if val == list[i] {
			return true
		}
	}
	return false
}

func HandlerRecomNewVideos() func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query().Get("index")
			choices := []string{"tv", "movie"}
			if !containList(choices, query) {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			cates := ""
			if query == "tv" {
				cates = "国产剧"
			}
			if query == "movie" {
				cates = "最新"
			}
			resutl, err := FetchDoubanMovie(query, cates)
			if err != nil {
				log.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			backJson := RecomList{
				TotalList: make([]RecomItem, 0),
			}
			for _, v := range resutl.Subjects {
				temp := RecomItem{
					Title: v.Title,
					Rate:  v.Rate,
					Image: v.Cover,
				}
				backJson.TotalList = append(backJson.TotalList, temp)
			}
			jsonBytes, err := json.Marshal(backJson)
			if err != nil {
				log.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(jsonBytes)
		},
	)
}
