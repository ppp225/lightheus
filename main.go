package main

import (
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"syscall"

	"github.com/ppp225/aetos"
	"github.com/ppp225/go-common"

	"github.com/go-playground/validator"
	"gopkg.in/yaml.v2"
)

type config struct {
	Addresses   []string `yaml:"addresses" validate:"required"`
	Address     string   `yaml:"address" validate:""`
	MetricsPath string   `yaml:"metrics_path" validate:""`
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

func parseUrlsLAndGetAetosFiles(urls []string) []aetos.File {
	af := make([]aetos.File, 0, len(urls))
	for _, a := range urls {

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

		f := aetos.File{
			FilePath: "somepath",
			Labels:   map[string]string{"host": h, "path": p, "scheme": "mobile"},
		}
		af = append(af, f)
	}
	common.PrettyPrint(af)
	return af
}

// Lightheus represents Lightheus instance
type Lightheus struct {
	promExporter *aetos.Aetos
}

// New creates new Lightheus instance
func New(configPath string) *Lightheus {
	cfg := getConfigFromFile(configPath)
	af := parseUrlsLAndGetAetosFiles(cfg.Addresses)

	ae := aetos.NewBaseWithFiles("aetos-base.yml", af)
	return &Lightheus{
		promExporter: ae,
	}
}

func (v *Lightheus) Run() {

}

func main() {

	binary, lookErr := exec.LookPath("lighthouse")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"lighthouse", "--chrome-flags=--headless --no-sandbox", "--no-enable-error-reporting", "http://google.com", "--output", "json", "--output-path", "./google.json"}

	env := os.Environ()

	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}

	// lh := New("light2.yml")
	// lh.Run()
}
