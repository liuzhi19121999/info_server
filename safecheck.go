package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type CodeStorages struct {
	CodesList []string `json:"codes"`
}

type SafeCodes struct {
	NeedSafe   bool         `json:"safety"`
	CodesInfos CodeStorages `json:"codeinfo"`
	StartTime  int64        `json:"time"`
}

var codeStorages SafeCodes

const codesPath = "safeCode.json"

func initSafeCode() {
	var codeStorage CodeStorages
	jsonBytes, err := os.ReadFile(codesPath)
	if err == nil {
		err = json.Unmarshal(jsonBytes, &codeStorage)
		if err != nil {
			log.Println(err.Error())
			codeStorage = CodeStorages{
				CodesList: make([]string, 0),
			}
		}
	} else {
		codeStorage = CodeStorages{
			CodesList: make([]string, 0),
		}
	}
	codeStorages.CodesInfos = codeStorage
	codeStorages.NeedSafe = true
	codeStorages.StartTime = time.Now().Unix()
	if len(codeStorage.CodesList) == 0 {
		codeStorages.NeedSafe = false
	}
}

func (s SafeCodes) refreshCodes() {
	nowTime := time.Now().Unix()
	if (nowTime - s.StartTime) >= 300 {
		fileByte, err := os.ReadFile(codesPath)
		if err != nil {
			log.Println(err.Error())
			return
		}
		err = json.Unmarshal(fileByte, &codeStorages.CodesInfos)
		if err != nil {
			log.Println(err.Error())
			return
		}
		if len(codeStorages.CodesInfos.CodesList) == 0 {
			codeStorages.StartTime = time.Now().Unix()
			codeStorages.NeedSafe = false
			return
		}
		codeStorages.StartTime = time.Now().Unix()
		codeStorages.NeedSafe = true
	}
}

func (s SafeCodes) checkIsAvaliable(code string) bool {
	for _, v := range s.CodesInfos.CodesList {
		if v == code {
			return true
		}
	}
	return false
}

func netMiddleWare(handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			tokens := r.URL.Query().Get("tokens")
			codeStorages.refreshCodes()
			if codeStorages.NeedSafe {
				if !codeStorages.checkIsAvaliable(tokens) {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
			}
			handler.ServeHTTP(w, r)
		},
	)
}

type SecretCodeStruct struct {
	CodeList   []string `json:"codelist"`
	CodeHash   string   `json:"hash"`
	SecretInt  []int    `json:"secretNum"`
	UseEnigma  bool     `json:"useable"`
	TimeStramp int64    `json:"-"`
}

const secretCodePath = "secret.json"

var secretCodeList SecretCodeStruct

func initSecretCode() {
	fileBytes, err := os.ReadFile(secretCodePath)
	if err != nil {
		log.Println(err.Error())
		return
	}
	err = json.Unmarshal(fileBytes, &secretCodeList)
	if err != nil {
		log.Println(err.Error())
		secretCodeList = SecretCodeStruct{
			CodeList:   make([]string, 0),
			CodeHash:   "",
			SecretInt:  make([]int, 0),
			UseEnigma:  false,
			TimeStramp: 0,
		}
	}
	secretCodeList.TimeStramp = time.Now().Unix()
}

func (s SecretCodeStruct) refreshData() {
	timeNow := time.Now().Unix()
	if (timeNow - s.TimeStramp) >= 7200 {
		fileBytes, err := os.ReadFile(secretCodePath)
		if err != nil {
			log.Println(err.Error())
			return
		}
		err = json.Unmarshal(fileBytes, &secretCodeList)
		if err != nil {
			log.Println(err.Error())
			secretCodeList = SecretCodeStruct{
				CodeList:   make([]string, 0),
				CodeHash:   "",
				SecretInt:  make([]int, 0),
				UseEnigma:  false,
				TimeStramp: 0,
			}
		}
		secretCodeList.TimeStramp = time.Now().Unix()
	}
}

func HandlerTransformHashCode() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			secretCodeList.refreshData()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(secretCodeList.CodeHash))
		},
	)
}

type CodeList struct {
	ListCodes []string `json:"listcodes"`
}

func HandlerTransCodeList() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var codes CodeList
			codes.ListCodes = secretCodeList.CodeList
			jsonBytes, err := json.Marshal(codes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err.Error())
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(jsonBytes)
		},
	)
}

func getIndex(item string, strList []string) int {
	for i := 0; i < len(strList); i++ {
		if strList[i] == item {
			return i
		}
	}
	return -1
}

func encodeds(inputJson string) string {
	var codeInt = make([]int, len(secretCodeList.SecretInt))
	copy(codeInt, secretCodeList.SecretInt)
	strList := secretCodeList.CodeList
	totalLength := len(strList)
	inputList := strings.Split(inputJson, "")
	inputLength := len(inputList)
	res := ""
	//fmt.Println(codeInt)
	for i := 0; i < len(inputList); i += 5 {
		for j := 0; j < 5; j++ {
			if i+j >= inputLength {
				break
			}
			tempIndex := getIndex(inputList[i+j], strList)
			if tempIndex == -1 {
				res += inputList[i+j]
				//fmt.Printf(inputList[i+j] + " ")
			} else {
				tempIndex += codeInt[j]
				for {
					if tempIndex < totalLength {
						break
					}
					tempIndex -= totalLength
				}
				res += strList[tempIndex]
				//fmt.Println(inputList[i+j] + " -> " + strList[tempIndex])
			}
			codeInt[j] = codeInt[j] + 1
			for {
				if codeInt[j] < totalLength {
					break
				}
				codeInt[j] = codeInt[j] - totalLength
			}
		}
	}
	return res
}
