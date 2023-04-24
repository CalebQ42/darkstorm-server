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
