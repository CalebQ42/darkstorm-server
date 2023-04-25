package darkstormtech

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/CalebQ42/stupid-backend"
	"github.com/CalebQ42/stupid-backend/pkg/defaultapp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DarkstormTech struct {
	*defaultapp.App
	filesFolder string
}

func NewDarkstormTech(c *mongo.Client, filesFolder string) *DarkstormTech {
	return &DarkstormTech{
		App:         defaultapp.NewDefaultApp(c.Database("darkstormtech")),
		filesFolder: filesFolder,
	}
}

func (d *DarkstormTech) Extension(req *stupid.Request) bool {
	if req.Path[0] != "page" {
		return false
	}
	if len(req.Path) == 1 {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	}
	if req.Path[1] == "files" {
		return d.handleFiles(req)
	} else if req.Path[1] == "resume" {
		return d.handleResume(req)
	}
	res := d.DB.Collection("pages").FindOne(context.TODO(), bson.M{"_id": strings.Join(req.Path[1:], "/")}, options.FindOne().SetProjection(bson.M{"_id": 0, "content": 1}))
	if res.Err() == mongo.ErrNoDocuments {
		req.Resp.WriteHeader(http.StatusNotFound) //TODO: Give some sort of default page.
		return true
	} else if res.Err() != nil {
		log.Println("Error while getting page:", res.Err())
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	pag := struct { //TODO: Add favicon and title support.
		Content string
	}{}
	err := res.Decode(&pag)
	if err != nil {
		log.Println("Error while decoding page:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	_, err = req.Resp.Write([]byte(pag.Content))
	if err != nil {
		log.Println("Error while writing response:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
	}
	return true
}

func (d *DarkstormTech) handleFiles(req *stupid.Request) bool {
	fils, err := os.ReadDir(d.filesFolder)
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
	_, err = req.Resp.Write([]byte(out))
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

func (d *DarkstormTech) handleResume(req *stupid.Request) bool {
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
		out += "<h2 style='margin-bottom:10px'>" + p.ID + "</h2>"
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
	_, err = req.Resp.Write([]byte(out))
	if err != nil {
		log.Println("Error while writing output:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
	}
	return true
}
