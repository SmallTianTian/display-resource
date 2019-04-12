package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	Cache DataList
	index = func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./index.html")
	}
	reset = func(w http.ResponseWriter, r *http.Request) {
		b, e := ioutil.ReadFile("./data.json")
		if e != nil {
			w.WriteHeader(500)
		}
		e = json.Unmarshal(b, &Cache)
		if e != nil {
			w.WriteHeader(500)
		}
		index(w, r)
	}
	findFunc = func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		id, _ := strconv.Atoi(vars["id"][0])
		result := find(id)
		b, _ := json.Marshal(result)
		w.Write(b)
	}
	queryFunc = func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		nameTmp := vars["name"]
		fromTmp := vars["from"]
		var name, from string
		if len(nameTmp) > 0 {
			name = nameTmp[0]
		}
		if len(fromTmp) > 0 {
			from = fromTmp[0]
		}
		result := query(from, name)
		b, _ := json.Marshal(result)
		w.Write(b)
	}
	banner = func(w http.ResponseWriter, r *http.Request) {
		dirs, _ := ioutil.ReadDir("./")

		banners := make([]string, 0)
		for _, f := range dirs {
			if !f.IsDir() && strings.HasPrefix(f.Name(), "banner") {
				banners = append(banners, f.Name())
			}
		}

		sort.Strings(banners)

		b, _ := json.Marshal(banners)
		w.Write(b)
	}
)

type DataList []*Data

func (d DataList) Filter(c func(*Data) bool) (result []*Data) {
	for _, item := range Cache {
		if c(item) {
			result = append(result, item)
		}
	}
	return
}

type ItemAbout struct {
	Id  int    `json:"id"`
	Img string `json:"img"`
}

type itemdetail struct {
	Img         string       `json:"img"`
	Name        string       `json:"name"`
	Author      string       `json:"author"`
	Type        []string     `json:"type"`
	Status      string       `json:"status"`
	Price       int          `json:"price"`
	Describe    string       `json:"describe"`
	Correlation []*ItemAbout `json:"correlation"`
}

type Data struct {
	ID       int      `json:"id"`
	Img      string   `json:"img"`
	Eye      int      `json:"eye"`
	Heart    int      `json:"heart"`
	Download int      `json:"download"`
	Name     string   `json:"name"`
	Author   string   `json:"author"`
	Type     []string `json:"type"`
	Price    int      `json:"price"`
	Describe string   `json:"describe"`
	Status   string   `json:"status"`
}

func init() {
	b, e := ioutil.ReadFile("./data.json")
	if e != nil {
		fmt.Println("读取文件异常。", e)
		os.Exit(0)
	}
	e = json.Unmarshal(b, &Cache)
	if e != nil {
		fmt.Println("读取文件异常。", e)
		os.Exit(0)
	}
	fmt.Println("程序数据加载成功。")
}

func getStaticFile(mux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(strings.Split(r.RequestURI, ".")) > 1 {
			http.ServeFile(w, r, r.RequestURI[1:])
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func query(from, name string) (result []*Data) {
	if from == "" {
		return Cache
	}

	t := func(d *Data) bool {
		switch from {
		case "type":
			for _, i := range d.Type {
				if i == name {
					return true
				}
			}
			return false
		case "author":
			return d.Author == name
		case "name":
			return strings.HasPrefix(d.Name, name)
		}
		return false
	}
	result = Cache.Filter(t)
	return
}

func find(id int) (result *itemdetail) {
	for _, item := range Cache {
		if item.ID == id {
			result = &itemdetail{
				Img:         item.Img,
				Name:        item.Name,
				Author:      item.Author,
				Type:        item.Type,
				Status:      item.Status,
				Price:       item.Price,
				Describe:    item.Describe,
				Correlation: make([]*ItemAbout, 0),
			}
		}
	}
	for _, item := range Cache {
		if item.Author == result.Author && item.ID != id {
			about := &ItemAbout{Id: item.ID, Img: item.Img}
			result.Correlation = append(result.Correlation, about)
		}
	}
	// TODO 加类型
	return
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/index", index)
	http.HandleFunc("/index.html", index)

	http.HandleFunc("/query", queryFunc)
	http.HandleFunc("/detail", findFunc)
	http.HandleFunc("/reset", reset)
	http.HandleFunc("/banner_list", banner)

	server := &http.Server{Addr: ":8080", Handler: getStaticFile(http.DefaultServeMux)}
	server.ListenAndServe()
}
