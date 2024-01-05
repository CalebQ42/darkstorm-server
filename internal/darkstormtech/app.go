package darkstormtech

import (
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/CalebQ42/bbConvert"
	"github.com/CalebQ42/stupid-backend/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DarkstormTech struct {
	stupid.UnKeyedApp
	bb          *bbConvert.HTMLConverter
	DB          *mongo.Database
	filesFolder string
}

func NewDarkstormTech(c *mongo.Client, filesFolder string) *DarkstormTech {
	bb := &bbConvert.HTMLConverter{}
	bb.ImplementDefaults()
	return &DarkstormTech{
		bb:          bb,
		DB:          c.Database("darkstormtech"),
		filesFolder: filesFolder,
	}
}

func (d *DarkstormTech) AlternateName() string {
	return "page"
}

type pageOut struct {
	Content string `json:"content"`
	Title   string `json:"title"`
	Favicon string `json:"favicon"`
}

func notFoundPage() pageOut {
	return pageOut{
		Content: "404 Page Not Found 😥",
		Title:   "Darkstorm.Tech",
		Favicon: "https://darkstorm.tech/favicon.png",
	}
}

func pageWith(content string, title string) pageOut {
	if title == "" {
		title = "Darkstorm.Tech"
	}
	return pageOut{
		Content: content,
		Title:   title,
		Favicon: "https://darkstorm.tech/favicon.png",
	}
}

func (p pageOut) json() []byte {
	out, _ := json.Marshal(p)
	return out
}

func (p *pageOut) addDefaults() {
	if p.Title == "" {
		p.Title = "Darkstorm.Tech"
	}
	if p.Favicon == "" {
		p.Favicon = "https://darkstorm.tech/favicon.png"
	}
}

func (d *DarkstormTech) HandleReqest(req *stupid.Request) bool {
	if req.Path[0] != "page" {
		return false
	}
	if len(req.Path) == 1 {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	}
	switch req.Path[1] {
	case "blog":
		return d.handleBlog(req)
	case "files":
		return d.handleFiles(req)
	case "portfolio":
		return d.handlePortfolio(req)
	}
	res := d.DB.Collection("pages").FindOne(context.TODO(), bson.M{"_id": strings.Join(req.Path[1:], "/")}, options.FindOne().SetProjection(bson.M{"_id": 0}))
	if res.Err() == mongo.ErrNoDocuments {
		req.Resp.Write(notFoundPage().json())
		req.Resp.WriteHeader(http.StatusNotFound)
		return true
	} else if res.Err() != nil {
		log.Println("Error while getting page:", res.Err())
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	var pag pageOut
	err := res.Decode(&pag)
	if err != nil {
		log.Println("Error while decoding page:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	pag.Content = d.bb.Convert(pag.Content)
	(&pag).addDefaults()
	_, err = req.Resp.Write(pag.json())
	if err != nil {
		log.Println("Error while writing response:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
	}
	return true
}

type blog struct {
	ID      string `bson:"_id" json:"id"`
	Title   string `bson:"title" json:"title"`
	Content string `bson:"content" json:"content"`
}

type blogOut struct {
	Content string `json:"content"`
	Title   string `json:"title"`
}

func (b blogOut) json() []byte {
	out, _ := json.Marshal(b)
	return out
}

func (d *DarkstormTech) handleBlog(req *stupid.Request) bool {
	if req.Method == http.MethodPost {
		return d.addBlog(req)
	} else if req.Method != http.MethodGet {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	}
	var res *mongo.SingleResult
	if len(req.Path) == 2 {
		res = d.DB.Collection("blogs").FindOne(context.TODO(), bson.M{}, options.FindOne().SetSort(bson.M{"_id": -1}))
	} else {
		res = d.DB.Collection("blogs").FindOne(context.TODO(), bson.M{"_id": req.Path[2]})
	}
	if res.Err() == mongo.ErrNoDocuments {
		req.Resp.Write(notFoundPage().json())
		req.Resp.WriteHeader(http.StatusNotFound)
		return true
	} else if res.Err() != nil {
		log.Println("Error while getting blog:", res.Err())
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	var b blogOut
	err := res.Decode(&b)
	if err != nil {
		log.Println("Error while decoding blog:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	b.Content = d.bb.Convert(b.Content)
	_, err = req.Resp.Write(b.json())
	if err != nil {
		log.Println("Error while writing response:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
	}
	return true
}

func (d *DarkstormTech) addBlog(req *stupid.Request) bool {
	if req.User == nil {
		req.Resp.WriteHeader(http.StatusUnauthorized)
		return true
	}
	//TODO: Check if user is admin
	if req.Body == nil {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	}
	bod, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		log.Println("Error while reading body:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	if len(bod) == 0 {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	}
	var b blog
	err = json.Unmarshal(bod, &b)
	if err != nil {
		log.Println("Error while unmarshalling body:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	b.ID = strconv.Itoa(int(time.Now().Unix()))
	_, err = d.DB.Collection("blogs").InsertOne(context.TODO(), b)
	if err != nil {
		log.Println("Error while inserting blog:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	return true
}

func (d *DarkstormTech) handleFiles(req *stupid.Request) bool {
	if req.Method != http.MethodGet {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	}
	foldPath := ""
	if len(req.Path) > 1 {
		foldPath = filepath.Join(req.Path[2:]...)
	}
	fils, err := os.ReadDir(filepath.Join(d.filesFolder, foldPath))
	if err != nil {
		log.Println("Error while getting files:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	out := ""
	var inf fs.FileInfo
	for _, f := range fils {
		if f.IsDir() {
			continue
		}
		inf, err = f.Info()
		if err != nil {
			log.Println("Error while getting FileInfo for", f.Name(), err)
			req.Resp.WriteHeader(http.StatusInternalServerError)
			return true
		}
		out += "<p><a href='https://darkstorm.tech/files/" + f.Name() + "'>" + f.Name() + "</a> " + inf.ModTime().Round(time.Minute).String() + "</p>\n"
	}
	_, err = req.Resp.Write(pageWith(out, "Files").json())
	if err != nil {
		log.Println("Error while writing output:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
	}
	return true
}

type project struct {
	ID          string `bson:"_id"`
	Repository  string
	Description string
	Language    []struct {
		Language string
		Dates    string
	}
}

func selectedString(selected bool) string {
	if !selected {
		return ""
	}
	return " selected"
}

func (d *DarkstormTech) handlePortfolio(req *stupid.Request) bool {
	if req.Method != http.MethodGet {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	}
	filter := bson.M{}
	lang := ""
	if l, ok := req.Query["lang"]; ok && len(l) == 1 && l[0] != "" {
		lang = l[0]
		filter = bson.M{"language.language": l[0]}
	}
	projects := make([]project, 0)
	res, err := d.DB.Collection("projects").Find(context.TODO(), filter)
	if err != nil {
		log.Println("Error while getting projects:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	err = res.All(context.TODO(), &projects)
	if err != nil {
		log.Println("Error while decoding projects:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	out := "<p>Language Filter: <select name='langSelect' id='langSelect'>"
	out += "<option value=''" + selectedString(lang == "") + ">All</option>"
	out += "<option value='Go'" + selectedString(lang == "Go") + ">Go</option>"
	out += "<option value='Dart'" + selectedString(lang == "Dart") + ">Dart (Flutter)</option>"
	out += "<option value='Java'" + selectedString(lang == "Java") + ">Java</option>"
	out += "</select></p>"
	for _, p := range projects {
		out += "<h1 style='margin-bottom:10px'>" + p.ID + "</h1>"
		out += "<p><a href='" + p.Repository + "'>" + p.Repository + "</a></p>"
		for _, l := range p.Language {
			lang := l.Language
			if lang == "Dart" {
				lang = "Dart (Flutter)"
			}
			out += "<p><b>" + lang + "</b>: " + l.Dates + "</p>"
		}
		out += "<p>" + p.Description + "</p>"
	}
	_, err = req.Resp.Write(pageWith(out, "Portfolio").json())
	if err != nil {
		log.Println("Error while writing output:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
	}
	return true
}
