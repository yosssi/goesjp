package main

import (
	"fmt"
	"github.com/drone/routes"
	"github.com/eknkc/amber"
	"github.com/yosssi/goesjp/consts"
	"github.com/yosssi/gologger"
	"github.com/yosssi/goutils"
	"html/template"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"launchpad.net/goyaml"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

type Link struct {
	Id           bson.ObjectId `bson:"_id"`
	Url          string        `bson:"Url"`
	Title        string        `bson:"Title"`
	ErrorMessage string        `bson:"ErrorMessage"`
	CreatedAt    time.Time     `bson:"CreatedAt"`
	UpdatedAt    time.Time     `bson:"UpdatedAt"`
}

var (
	loggerYaml map[string]string = make(map[string]string)
	serverYaml map[string]string = make(map[string]string)
	mgoYaml    map[string]string = make(map[string]string)
	logger     gologger.Logger
	version    string                        = strings.TrimLeft(runtime.Version(), "go")
	templates  map[string]*template.Template = make(map[string]*template.Template)
)

func main() {
	initialize()
	serve()
}

func initialize() {
	print("initialize starts.")
	setYaml("loggerYaml", consts.LoggerYamlPath, &loggerYaml)
	setLogger()
	setYaml("serverYaml", consts.ServerYamlPath, &serverYaml)
	setYaml("mgoYaml", consts.MgoYamlPath, &mgoYaml)
	logger.Info("initialize ends.")
}

func serve() {
	pwd, _ := os.Getwd()
	mux := routes.New()
	if isDebug() {
		mux.Static("/", pwd)
	} else {
		mux.Static("/public", pwd)
	}
	mux.Get("/", top)

	http.Handle("/", mux)
	logger.Info("Listening on port", serverYaml["Port"])
	http.ListenAndServe(":"+serverYaml["Port"], nil)
}

func setYaml(name string, filePath string, yaml *map[string]string) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	err = goyaml.Unmarshal(bytes, yaml)
	if err != nil {
		panic(err)
	}
	if logger.Name != "" {
		logger.Info(name, "was set.", "yaml:", *yaml)
	} else {
		print(name, "was set.", "yaml:", *yaml)
	}
}

func setLogger() {
	logger = gologger.GetLogger(loggerYaml)
	logger.Info("logger was set.", "logger:", goutils.StructToMap(&logger))
}

func print(s ...interface{}) {
	fmt.Print(now(), " - ")
	fmt.Println(s...)
}

func now() string {
	return time.Now().Format(consts.TimeFormatLayout)
}

func top(w http.ResponseWriter, r *http.Request) {
	session, err := mgo.Dial(mgoYaml["Host"])
	if err != nil {
		handleError(w, err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB(mgoYaml["Db"]).C("links")
	links := make([]Link, 0)
	err = c.Find(bson.M{"Title": bson.M{"$ne": ""}, "ErrorMessage": ""}).Sort("-UpdatedAt").All(&links)
	render(w, "./views/top.amber", map[string]interface{}{
		"IsDebug": isDebug(),
		"Links":   links,
		"Version": version,
	})
}

func handleError(w http.ResponseWriter, err error) {
	logger.Error(err.Error())
	http.Error(w, consts.ErrorMessageInternalServerError, http.StatusInternalServerError)
}

func render(w http.ResponseWriter, file string, data interface{}) {
	if !isDebug() {
		tpl, prs := templates[file]
		if prs {
			tpl.Execute(w, data)
			return
		}
	}
	compiler := amber.New()
	err := compiler.ParseFile(file)
	if err != nil {
		handleError(w, err)
	}
	tpl, err := compiler.Compile()
	if err != nil {
		handleError(w, err)
	}
	if !isDebug() {
		templates[file] = tpl
	}
	tpl.Execute(w, data)
}

func isDebug() bool {
	return serverYaml["Debug"] == "true"
}
