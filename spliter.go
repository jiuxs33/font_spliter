package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"

	"github.com/golang/freetype/truetype"

	"github.com/go-courier/fontnik"
)

type ResultHandler struct {
	ttf        []byte
	logger     *log.Logger
	builder    *fontnik.SDFBuilder
	fontGlyhps [] *fontnik.Glyph
}

func (h *ResultHandler) Init() {
	var err error
	h.ttf, err = ioutil.ReadFile("./fonts/PingFang SC Semibold.ttf")
	if err != nil {
		panic(err)
	}
	font, fontErr := truetype.Parse(h.ttf)
	if fontErr != nil {
		panic(err)
	}

	h.fontGlyhps = make([]*fontnik.Glyph, 65536);

	h.builder = fontnik.NewSDFBuilder(font, fontnik.SDFBuilderOpt{FontSize: 24, Buffer: 3})

	for i:=0; i<65536; i++ {
		g := h.builder.Glyph(rune(i))
		h.fontGlyhps[i] = g
	}

	file, err := os.OpenFile("spliter.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		log.Fatalln("fail to create spliter.log file!")
	}
	h.logger = log.New(file, "", log.Llongfile)
	h.logger.SetFlags(log.LstdFlags) // 设置写入文件的log日志的格式
}

func (h *ResultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin","*")
	//fmt.Println(r.URL.Path)
	//fmt.Println(r.URL.RawPath)
	path := r.URL.Path
	if strings.HasSuffix(path, ".pbf") {
		path = strings.TrimSuffix(path, ".pbf")
		pathParts := strings.Split(strings.TrimPrefix(path, "/"), "/")
		//fmt.Println(len(pathParts))
		//fmt.Println(pathParts)
		if len(pathParts) == 2 {
			ids := pathParts[1]
			idsArr := strings.Split(ids, ",")
			idsLen := len(idsArr)
			if idsLen > 0 {
				idsIntArr := make([]int, idsLen)
				idx := 0
				temp := 0
				for _, id := range idsArr {
					idInt, err := strconv.Atoi(id)
					if err == nil {
						idsIntArr[idx] = temp + idInt
						temp = idsIntArr[idx];
						idx++
					}
				}
				temp = 0
				//fmt.Println(idsIntArr)
				min := idsIntArr[0]
				max := idsIntArr[len(idsIntArr)-1]
				rng := fmt.Sprintf("%d-%d", min, max)
				fontFamily := h.builder.Font.Name(truetype.NameIDFontFullName)

				stack := &fontnik.Fontstack{}
				stack.Range = &rng
				stack.Name = &fontFamily
				for _, id := range idsIntArr {
					//fmt.Println(id)
					g := h.fontGlyhps[id]
					if g != nil {
						stack.Glyphs = append(stack.Glyphs, g)
					}
				}

				fontSplited := &fontnik.Glyphs{Stacks: []*fontnik.Fontstack{stack}}

				bytes, err := proto.Marshal(fontSplited)
				if err != nil {
					fmt.Println("error in Marshal")
				}
				w.Write(bytes)

			} else {
				fmt.Fprintln(w, "wrong parameters")
			}
		} else {
			fmt.Fprintln(w, "invalid request URL")
		}
	} else {
		fmt.Fprintln(w, "invalid request URL")
	}

}

func main() {
	params := os.Args
	port := "8088"
	if len(params)>1 {
		port = params[1]
	}
	handler := ResultHandler{}
	handler.Init()
	server := http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: &handler,
	}

	server.ListenAndServe()
}
