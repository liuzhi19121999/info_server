package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/anaskhan96/soup"
)

const fetchTimeOutSet = time.Second * 15
const infosTimeOutSet = time.Second * 16

const urlpath string = "http://cj.com/index.php/vod/search.html?wd="

const linagziBasic string = "http://cj.com"

type SearchResult struct {
	Name    string `json:"name"`
	Urlpath string `json:"path"`
	State   string `json:"state"`
}

type SourceData struct {
	Station string         `json:"station"`
	Infos   []SearchResult `json:"infos"`
}

type ConbinedSearchData struct {
	Hosts   []string   `json:"hosts"`
	FeiSu   SourceData `json:"feisu"`
	Liangzi SourceData `json:"liangzi"`
	GuangSu SourceData `json:"guangsu"`
	HonNiu  SourceData `json:"honniu"`
}

type BroadCastItem struct {
	M3u8Path string `json:"m3u8"`
	Label    string `json:"label"`
}

type BoradCastPage struct {
	Title     string          `json:"title"`
	Actors    string          `json:"actors"`
	ImagePath string          `json:"imgpath"`
	Year      string          `json:"year"`
	Contents  string          `json:"content"`
	Items     []BroadCastItem `json:"items"`
}

// unknown
func LZGetHtml(name string) (SourceData, error) {
	liangziData := SourceData{
		Station: "unkonwn",
		Infos:   make([]SearchResult, 0),
	}
	resp, err := soup.GetWithClient(urlpath+name,
		&http.Client{Timeout: time.Duration(fetchTimeOutSet)})
	if err != nil {
		log.Println(err.Error())
		return liangziData, err
	}
	parsed := soup.HTMLParse(resp)
	res := parsed.FindAll("a", "class", "videoName")
	if len(res) == 0 {
		return liangziData, errors.New("数据为空")
	}
	for _, v := range res {
		state := v.Find("i")
		stateText := ""
		if state.Error == nil {
			stateText = state.Text()
		}

		tempItem := SearchResult{
			Name:    v.Text(),
			Urlpath: v.Attrs()["href"],
			State:   stateText,
		}
		liangziData.Infos = append(liangziData.Infos, tempItem)
	}
	return liangziData, nil
}

func LZGetM3U8Data(urlpath string) (BoradCastPage, error) {
	broadcast := BoradCastPage{
		Title:     "",
		Actors:    "",
		ImagePath: "",
		Year:      "",
		Contents:  "",
	}
	m3u8List := make([]BroadCastItem, 0)
	response, err := soup.GetWithClient(linagziBasic+urlpath,
		&http.Client{Timeout: time.Duration(infosTimeOutSet)})
	if err != nil {
		log.Println(err.Error())
		return broadcast, err
	}
	parserData := soup.HTMLParse(response)
	// 信息栏
	infosBlock := parserData.Find("div", "class", "right")
	if infosBlock.Error == nil {
		infos := infosBlock.FindAll("p")
		if len(infos) >= 8 {
			broadcast.Title = infos[0].Text()
			broadcast.Actors = infos[6].Text()
			broadcast.Year = infos[7].Text()
		}
	}
	// 图片地址
	imgBlock := parserData.Find("div", "class", "left")
	if imgBlock.Error == nil {
		imgPathItem := imgBlock.Find("img")
		if imgPathItem.Error == nil {
			broadcast.ImagePath = imgPathItem.Attrs()["src"]
		}
	}
	contentBlock := parserData.Find("div", "class", "vod_content")
	if contentBlock.Error == nil {
		textBlock := contentBlock.Find("p")
		if textBlock.Error == nil {
			broadcast.Contents = textBlock.Text()
		}
	}
	m3u8Content := parserData.FindAll("input", "name", "copy_lzm3u8[]")
	if len(m3u8Content) == 0 {
		broadcast.Items = m3u8List
		return broadcast, errors.New("数据为空")
	}
	for _, v := range m3u8Content {
		infos := strings.Split(v.Attrs()["value"], "$")
		if len(infos) != 2 {
			continue
		}
		tempItem := BroadCastItem{
			Label:    infos[0],
			M3u8Path: infos[1],
		}
		m3u8List = append(m3u8List, tempItem)
	}
	broadcast.Items = m3u8List
	return broadcast, nil
}

// unkonwn
var feisuyun string = "https://f.com/vod/search/?wd="
var feisuyunBasic string = "https://f.com/"

func FSGetData(name string) (SourceData, error) {
	//changefeisuString()
	sourceData := SourceData{
		Station: "unknown",
		Infos:   make([]SearchResult, 0),
	}
	response, err := soup.GetWithClient(feisuyun+name,
		&http.Client{Timeout: time.Duration(fetchTimeOutSet)})
	if err != nil {
		log.Println(err.Error())
		return sourceData, err
	}
	parsedData := soup.HTMLParse(response)
	titles := parsedData.FindAll("li", "class", "clearfix")
	if len(titles) == 0 {
		return sourceData, errors.New("数据为空")
	}
	for _, v := range titles {
		aBlock := v.Find("a")
		if aBlock.Error != nil {
			continue
		}
		stateString := ""
		emBlock := aBlock.Find("em")
		if emBlock.Error == nil {
			stateString = emBlock.Text()
		}
		temp := SearchResult{
			Name:    aBlock.Text(),
			Urlpath: aBlock.Attrs()["href"],
			State:   stateString,
		}
		sourceData.Infos = append(sourceData.Infos, temp)
	}
	return sourceData, nil
}

func FSfetchM3U8Data(urlpath string) (BoradCastPage, error) {
	broadPage := BoradCastPage{
		Title:     "",
		Actors:    "",
		ImagePath: "",
		Year:      "",
		Contents:  "",
	}
	tempItems := make([]BroadCastItem, 0)
	response, err := soup.GetWithClient(feisuyunBasic+urlpath,
		&http.Client{Timeout: time.Duration(infosTimeOutSet)})
	if err != nil {
		return broadPage, err
	}
	parsedData := soup.HTMLParse(response)
	headerContent := parsedData.Find("div", "class", "stui-content__detail")
	if headerContent.Error == nil {
		titleBlock := headerContent.Find("h1", "class", "title")
		if titleBlock.Error == nil {
			broadPage.Title = titleBlock.Text()
		}
		infoBlocks := headerContent.FindAll("p")
		if len(infoBlocks) >= 5 {
			broadPage.Actors = infoBlocks[2].Text()
			yearItems := strings.Split(strings.ReplaceAll(infoBlocks[4].FullText(), "\n", ""), "：")
			if len(yearItems) >= 5 {
				broadPage.Year = yearItems[4]
			}
		}
	}
	contentBlock := parsedData.Find("div", "id", "desc")
	if contentBlock.Error == nil {
		stringList := contentBlock.FindAll("div")
		if len(stringList) >= 2 {
			broadPage.Contents = stringList[1].Text()
		}
	}
	imgBlock := parsedData.Find("img", "class", "img-responsive")
	if imgBlock.Error == nil {
		broadPage.ImagePath = imgBlock.Attrs()["src"]
	}
	broadPage.Items = tempItems
	m3u8Page := make([]soup.Root, 0)
	playListBlock := parsedData.FindAll("div", "id", "playlist")
	if len(playListBlock) >= 1 {
		m3u8Page = playListBlock[0].FindAll("li")
	}
	if len(m3u8Page) == 0 {
		return broadPage, errors.New("数据为空")
	}
	for _, v := range m3u8Page {
		aTagEle := v.Find("a", "class", "copy_text")
		if aTagEle.Error != nil {
			continue
		}
		tempm3u8Path := ""
		pathBlock := aTagEle.Find("span")
		if pathBlock.Error == nil {
			tempm3u8Path = strings.Replace(pathBlock.FullText(), "$", "", 1)
		}
		tempItem := BroadCastItem{
			Label:    aTagEle.Text(),
			M3u8Path: tempm3u8Path,
		}
		broadPage.Items = append(broadPage.Items, tempItem)
	}
	return broadPage, nil
}

// unknown
const guangsuyun string = "https://www.g.com/index.php/vod/search.html?wd="
const guangsuyunBasic string = "https://www.g.com/"

func GSGetData(name string) (SourceData, error) {
	guangsuData := SourceData{
		Station: "unknown",
		Infos:   make([]SearchResult, 0),
	}
	response, err := soup.GetWithClient(guangsuyun+name,
		&http.Client{Timeout: time.Duration(fetchTimeOutSet)})
	if err != nil {
		return guangsuData, err
	}
	parsed := soup.HTMLParse(response)
	tBodyBlock := parsed.Find("tbody")
	if tBodyBlock.Error != nil {
		return guangsuData, errors.New("解析错误")
	}
	results := tBodyBlock.FindAll("tr")
	if len(results) == 0 {
		return guangsuData, errors.New("数据为空")
	}
	for _, v := range results {
		aBlock := v.Find("a")
		if aBlock.Error != nil {
			continue
		}
		stateString := ""
		spanBlock := aBlock.Find("span")
		if spanBlock.Error == nil {
			stateString = spanBlock.Text()
		}
		searchTemp := SearchResult{
			Name:    strings.ReplaceAll(aBlock.Text(), " ", ""),
			Urlpath: aBlock.Attrs()["href"],
			State:   stateString,
		}
		guangsuData.Infos = append(guangsuData.Infos, searchTemp)
	}
	return guangsuData, nil
}

func GSGetM3U8Data(urlpath string) (BoradCastPage, error) {
	broadCastPage := BoradCastPage{
		Title:     "",
		Actors:    "",
		ImagePath: "",
		Year:      "",
		Contents:  "",
	}
	m3u8Items := make([]BroadCastItem, 0)
	response, err := soup.GetWithClient(guangsuyunBasic+urlpath,
		&http.Client{Timeout: time.Duration(infosTimeOutSet)})
	if err != nil {
		broadCastPage.Items = m3u8Items
		return broadCastPage, nil
	}
	parsed := soup.HTMLParse(response)
	imgBlock := parsed.Find("div", "class", "dy-photo")
	if imgBlock.Error == nil {
		imgB := imgBlock.Find("img")
		if imgB.Error == nil {
			broadCastPage.ImagePath = imgB.Attrs()["src"]
		}
	}
	infosBlock := parsed.Find("div", "class", "dy-deta")
	if infosBlock.Error == nil {
		infosList := infosBlock.FindAll("p")
		if len(infosList) >= 9 {
			broadCastPage.Title = infosList[0].Text()
			actorsList := strings.Split(infosList[8].Text(), "：")
			if len(actorsList) >= 2 {
				broadCastPage.Actors = actorsList[1]
			}
			yearList := strings.Split(infosList[4].Text(), "：")
			if len(yearList) >= 2 {
				broadCastPage.Year = yearList[1]
			}
		}
	}
	contentBlock := parsed.Find("p", "class", "dy-moreIns")
	if contentBlock.Error == nil {
		contetnsText := contentBlock.Text()
		contetnsText = strings.ReplaceAll(contetnsText, " ", "")
		contetnsText = strings.ReplaceAll(contetnsText, "\n", "")
		broadCastPage.Contents = contetnsText
	}
	m3u8ListConetent := make([]soup.Root, 0)
	playListBlock := parsed.FindAll("div", "class", "dy-collect-video")
	if len(playListBlock) >= 2 {
		m3u8ListConetent = playListBlock[1].FindAll("li")
	}
	if len(m3u8ListConetent) == 0 {
		broadCastPage.Items = m3u8Items
		return broadCastPage, errors.New("数据为空")
	}
	for _, v := range m3u8ListConetent {
		spanBlock := v.Find("span")
		if spanBlock.Error != nil {
			continue
		}
		aList := v.FindAll("a")
		if len(aList) == 0 {
			continue
		}
		tempM3U8Item := BroadCastItem{
			Label:    spanBlock.Text(),
			M3u8Path: aList[0].Text(),
		}
		m3u8Items = append(m3u8Items, tempM3U8Item)
	}
	broadCastPage.Items = m3u8Items
	return broadCastPage, nil
}

// unknown
const honniu string = "https://www.h.com/"

func HNGetData(name string) (SourceData, error) {
	fetch_url := "https://www.h.com/index.php/vod/search.html?wd=" + name + "&submit=search"
	HNData := SourceData{
		Station: "unknown",
		Infos:   make([]SearchResult, 0),
	}
	resp, err := soup.GetWithClient(fetch_url,
		&http.Client{Timeout: time.Duration(fetchTimeOutSet)})
	if err != nil {
		return HNData, err
	}
	parsed := soup.HTMLParse(resp)
	searchLists := parsed.FindAll("span", "class", "xing_vb4")
	if len(searchLists) == 0 {
		return HNData, errors.New("数据为空")
	}
	for _, v := range searchLists {
		tagA := v.Find("a")
		if tagA.Error != nil {
			continue
		}
		stateString := ""
		spanBlock := tagA.Find("em")
		if spanBlock.Error == nil {
			stateString = spanBlock.Text()
		}
		tempInfo := SearchResult{
			Name:    tagA.Text(),
			Urlpath: tagA.Attrs()["href"],
			State:   stateString,
		}
		HNData.Infos = append(HNData.Infos, tempInfo)
	}
	return HNData, nil
}

func HNM3U8Data(urlpath string) (BoradCastPage, error) {
	broadCast := BoradCastPage{
		Title:     "",
		Actors:    "",
		Year:      "",
		Contents:  "",
		ImagePath: "",
	}
	broadItems := make([]BroadCastItem, 0)
	resp, err := soup.GetWithClient(honniu+urlpath,
		&http.Client{Timeout: time.Duration(infosTimeOutSet)})
	if err != nil {
		broadCast.Items = broadItems
		return broadCast, err
	}
	parsed := soup.HTMLParse(resp)
	imgBlock := parsed.Find("img", "class", "lazy")
	if imgBlock.Error == nil {
		broadCast.ImagePath = imgBlock.Attrs()["href"]
	}
	titleBlock := parsed.Find("div", "class", "vodh")
	if titleBlock.Error == nil {
		h2Bolck := titleBlock.Find("h2")
		if h2Bolck.Error == nil {
			broadCast.Title = h2Bolck.Text()
		}
	}
	infoBlocks := parsed.Find("div", "class", "vodinfobox")
	if infoBlocks.Error == nil {
		infosList := infoBlocks.FindAll("li")
		if len(infosList) >= 7 {
			actorBlock := infosList[2].Find("span")
			if actorBlock.Error == nil {
				broadCast.Actors = actorBlock.Text()
			}
			yearBlock := infosList[6].Find("span")
			if yearBlock.Error == nil {
				broadCast.Year = yearBlock.Text()
			}
		}
	}
	contentBlock := parsed.Find("div", "class", "vodplayinfo")
	if contentBlock.Error == nil {
		broadCast.Contents = contentBlock.Text()
	}

	broadItesmBlock := make([]soup.Root, 0)
	m3u8Block := parsed.Find("div", "id", "play_1")
	if m3u8Block.Error == nil {
		ulBlock := m3u8Block.Find("ul")
		if ulBlock.Error == nil {
			broadItesmBlock = ulBlock.FindAll("li")
		}
	}
	for _, v := range broadItesmBlock {
		aTag := v.Find("a")
		if aTag.Error != nil {
			continue
		}
		tempItem := strings.Split(aTag.Text(), "$")
		if len(tempItem) != 2 {
			continue
		}
		temp := BroadCastItem{
			M3u8Path: tempItem[1],
			Label:    tempItem[0],
		}
		broadItems = append(broadItems, temp)
	}
	broadCast.Items = broadItems
	return broadCast, nil
}

// 服务器请求操作
func AllSearch(name string) ConbinedSearchData {
	var wg sync.WaitGroup
	conbinedData := ConbinedSearchData{
		Hosts:   []string{"feisu", "liangzi", "guangsu", "honniu"},
		Liangzi: SourceData{Station: "nujonwn", Infos: make([]SearchResult, 0)},
		FeiSu:   SourceData{Station: "unkonwn", Infos: make([]SearchResult, 0)},
		GuangSu: SourceData{Station: "unkonwn", Infos: make([]SearchResult, 0)},
		HonNiu:  SourceData{Station: "unknown", Infos: make([]SearchResult, 0)},
	}
	//fmt.Println(name)
	wg.Add(4)
	go func() {
		hnRes, err := HNGetData(name)
		if err != nil {
			conbinedData.HonNiu = hnRes
		}
		wg.Done()
	}()
	go func() {
		lzRes, err := LZGetHtml(name)
		if err == nil {
			conbinedData.Liangzi = lzRes
		}
		//fmt.Println("LZ OK")
		wg.Done()
	}()
	go func() {
		fsRes, err := FSGetData(name)
		if err == nil {
			conbinedData.FeiSu = fsRes
		}
		//fmt.Println("FS OK")
		wg.Done()
	}()
	go func() {
		gsRes, err := GSGetData(name)
		if err == nil {
			conbinedData.GuangSu = gsRes
		}
		//fmt.Println("GS OK")
		wg.Done()
	}()
	wg.Wait()
	return conbinedData
}

/*
type ReceptSearchData struct {
	Name string `json:"name"`
}
*/

func HandleSearchVideo() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			searchName := r.URL.Query().Get("name")
			searchCode := r.URL.Query().Get("wd")
			if searchCode != "submit" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			resData := AllSearch(searchName)
			jsonByte, err := json.Marshal(resData)
			if err != nil {
				log.Println(err.Error())
				w.WriteHeader(http.StatusExpectationFailed)
				return
			}
			if secretCodeList.UseEnigma {
				jsonString := string(jsonByte)
				res := encodeds(jsonString)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(res))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(jsonByte)
		},
	)
}

type ReceptBroadCastData struct {
	Path    string `json:"path"`
	Station string `json:"station"`
}

func HandleShowVideoPage() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			/*
				recpet := ReceptBroadCastData{}
				deocder := json.NewDecoder(r.Body)
				err := deocder.Decode(&recpet)
				if err != nil {
					log.Println(err.Error())
					w.WriteHeader(http.StatusNotAcceptable)
				}
			*/
			var err error
			stationName := r.URL.Query().Get("station")
			pathName := r.URL.Query().Get("path")
			recpet := ReceptBroadCastData{
				Station: stationName,
				Path:    pathName,
			}
			expectData := BoradCastPage{}
			if recpet.Station == "liangzi" {
				expectData, err = LZGetM3U8Data(recpet.Path)
				if err != nil {
					log.Println(err.Error())
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}
			if recpet.Station == "feisu" {
				expectData, err = FSfetchM3U8Data(recpet.Path)
				if err != nil {
					log.Println(err.Error())
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}
			if recpet.Station == "guangsu" {
				expectData, err = GSGetM3U8Data(recpet.Path)
				if err != nil {
					w.WriteHeader(http.StatusForbidden)
					log.Println(err.Error())
					return
				}
			}
			if recpet.Station == "honniu" {
				expectData, err = HNM3U8Data(recpet.Path)
				if err != nil {
					w.WriteHeader(http.StatusForbidden)
					log.Println(err.Error())
					return
				}
			}
			jsonByte, err := json.Marshal(expectData)
			if err != nil {
				log.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if secretCodeList.UseEnigma {
				jsonString := string(jsonByte)
				res := encodeds(jsonString)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(res))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(jsonByte)
		},
	)
}
