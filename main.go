package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/ppp225/aetos"
	"github.com/ppp225/go-common"

	"github.com/go-playground/validator"
	"gopkg.in/yaml.v2"
)

type config struct {
	Addresses []string `yaml:"addresses" validate:"required"`
}

func getConfigFromFile(path string) *config {
	ymlBytes := loadFile(path)
	var cfg config
	if err := yaml.Unmarshal(ymlBytes, &cfg); err != nil {
		log.Fatal(err)
	}
	validateConfig(&cfg)
	return &cfg
}

func loadFile(filename string) []byte {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	bytes, _ := ioutil.ReadAll(file)
	return bytes
}

func validateConfig(cfg *config) {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	if err := validate.Struct(cfg); err != nil {
		log.Fatal(err)
	}
}

func parseUrls(urls []string) (af []aetos.File, cc *crawlConfig) {
	af = make([]aetos.File, 0, len(urls))
	cc = &crawlConfig{
		file2urlMap: make(map[string]string),
	}
	for _, a := range urls {
		// prepare url
		if !strings.HasPrefix(a, "http://") && !strings.HasPrefix(a, "https://") {
			a = "https://" + a
		}
		u, err := url.Parse(a)
		if err != nil {
			panic(err)
		}
		h := u.Hostname()
		p := u.Path
		if p == "" {
			p = "/"
		}
		// parse url 2 filename
		r := strings.NewReplacer("/", "_", ":", "_", ".", "_")
		filename := r.Replace(a)
		//create objs
		f := aetos.File{
			FilePath: filename,
			Labels:   map[string]string{"host": h, "path": p, "scheme": "mobile"},
		}
		af = append(af, f)
		cc.file2urlMap[filename] = a
	}
	return af, cc
}

type crawlConfig struct {
	file2urlMap map[string]string
}

// Lightheus represents Lightheus instance
type Lightheus struct {
	cfg   *crawlConfig
	aetos *aetos.Aetos
}

// New creates new Lightheus instance
func New(configPath string) *Lightheus {
	cfg := getConfigFromFile(configPath)
	af, cc := parseUrls(cfg.Addresses)
	common.PrettyPrint(af)
	common.PrettyPrint(cc.file2urlMap)

	ae := aetos.NewBaseWithFiles("aetos-base.yml", af)
	ae.Debug()
	return &Lightheus{
		cfg:   cc,
		aetos: ae,
	}
}

func runLighthouse(outputJsonFile, url string) {
	log.Printf("Lightheus: Processing file=%q | url=%q", outputJsonFile, url)
	// cmd := exec.Command("lighthouse", "--chrome-flags=--headless --no-sandbox", "--no-enable-error-reporting", url, "--output", "json", "--output-path", outputJsonFile)
	cmd := exec.Command("lighthouse", "--chrome-flags=--headless --no-sandbox", "--no-enable-error-reporting", "--emulated-form-factor", "mobile", url, "--output", "json", "--output-path", outputJsonFile)
	stdout, err := cmd.Output()
	if err != nil {
		idx := bytes.LastIndex(stdout, []byte("\n"))
		if idx == -1 {
			log.Printf("Lightheus: Error err=%q, msg=%q", err, stdout)
		} else {
			// log.Printf("Lightheus: Error err=%q, msg=%q", err, stdout[idx:])
			log.Printf("Lightheus: Error err=%q, msg=%q", err, stdout)
		}
	}
}

func (v *Lightheus) Run() {
	log.Printf("Lightheus: Looping over %d entries:", len(v.cfg.file2urlMap))
	for file, url := range v.cfg.file2urlMap {
		runLighthouse(file, url)
	}

	go func() {
		for {
			for file, url := range v.cfg.file2urlMap {
				runLighthouse(file, url)
			}
			time.Sleep(time.Second * 10)
		}
	}()

	v.aetos.Run()
}

func main() {
	lh := New("lightheus.yml")
	lh.Run()
}
