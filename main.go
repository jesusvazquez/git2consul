package main

import (
	"github.com/thehivecorporation/log"
	"github.com/thehivecorporation/log/writers/text"
	"github.com/timjchin/unpuzzled"
	"gopkg.in/src-d/go-git.v4"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Consul Host
	ConsulHost string
	// Consul port
	ConsulPort int64
	// Local dir where to download the remote repository
	Directory string
	// RSA private key to access a private repository
	GitPrivateKey string
	// Git user to authenticate against your remote git provider
	GitUser string
	// Log Level
	LogLevel string
	// Polling interval to check for updates
	PollingInterval time.Duration
	// Git Repository address
	Repository string
	// Host address to start the router http server
	RouterHost string
	// Port that the router will bind to
	RouterPort int64
}

// Starts an http server that will serve a status endpoint
func startRouter(host *string, port *int64) {
	router := NewRouter()
	log.Info("Starting git2consul http server on " + *host + ":" + strconv.FormatInt(*port, 10) + " ...")
	log.Fatal(http.ListenAndServe(*host+":"+strconv.FormatInt(*port, 10), router).Error())
}

// Synchronous function that will check periodically if there are any changes in the remote repository
func gitLoop(config *Config) {
	// Clone github repository if it doesn't exist already
	cloneRepository(&config.Repository, &config.Directory, &config.GitUser, &config.GitPrivateKey)
	// Ticker periodicity
	tick := time.NewTicker(config.PollingInterval)
	for range tick.C {
		w := pullFromRepository(&config.Directory, &config.GitUser, &config.GitPrivateKey)
		iterateWorktree(w, config, "/")
	}
}

// Recursive function to iterates the git repository worktree and stores the
// key value pairs into consul KV
func iterateWorktree(worktree *git.Worktree, config *Config, parentDir string) {
	fs := worktree.Filesystem
	files, err := fs.ReadDir(parentDir)
	if err != nil {
		log.WithError(err).Fatal("Can't read directory" + parentDir)
		panic(err)
	}
	for _, file := range files {
		// If it is an iterable dir go deeper
		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
			iterateWorktree(worktree, config, parentDir+file.Name()+"/")
		} else if !strings.HasPrefix(file.Name(), ".") &&
			file.Name() != "README.md" {
			value, err := ioutil.ReadFile(config.Directory + parentDir + "/" + file.Name())
			if err != nil {
				log.WithError(err).Fatal("Can't read file" + config.Directory + parentDir + "/" + file.Name())
				panic(err)
			}
			putConsulKey(config, parentDir[1:len(parentDir)]+file.Name(), value)
		}
	}
	return
}

func main() {
	app := unpuzzled.NewApp()
	config := &Config{}
	app.OverridesOutputInTable = true
	app.Authors = []unpuzzled.Author{
		unpuzzled.Author{
			Name: "Jesus Vazquez",
		},
	}
	app.Command = &unpuzzled.Command{
		Name:  "git2consul",
		Usage: "Application that stores git content into consul key value.",
		Variables: []unpuzzled.Variable{
			&unpuzzled.StringVariable{
				Name:        "consul_host",
				Description: "Consul host address",
				Destination: &config.ConsulHost,
				Default:     "localhost",
			},
			&unpuzzled.Int64Variable{
				Name:        "consul_port",
				Description: "Consul port",
				Destination: &config.ConsulPort,
				Default:     8500,
			},
			&unpuzzled.StringVariable{
				Name:        "directory",
				Description: "Directory destination to clone git repository",
				Destination: &config.Directory,
				Default:     "/tmp/git2consul/repository",
			},
			&unpuzzled.StringVariable{
				Name:        "git_user",
				Description: "Git user to access a private repository",
				Destination: &config.GitUser,
				Default:     "git",
			},
			&unpuzzled.StringVariable{
				Name:        "git_private_key",
				Description: "Private key to access a private repository.",
				Destination: &config.GitPrivateKey,
			},
			&unpuzzled.DurationVariable{
				Name:        "polling_interval",
				Description: "Polling Interval duration",
				Destination: &config.PollingInterval,
				Default:     time.Minute,
			},
			&unpuzzled.StringVariable{
				Name:        "repository",
				Description: "Git source repository",
				Destination: &config.Repository,
				Required:    true,
			},
			&unpuzzled.StringVariable{
				Name:        "router_host",
				Description: "Specify the HOST",
				Destination: &config.RouterHost,
				Default:     "localhost",
			},
			&unpuzzled.Int64Variable{
				Name:        "router_port",
				Description: "Specify PORT.",
				Destination: &config.RouterPort,
				Default:     8090,
			},
		},
		Action: func() {
			// Set up logging
			log.SetWriter(text.New(os.Stdout))
			log.SetLevel(log.LevelInfo)

			// Run git polling loop
			go gitLoop(config)
			startRouter(&config.RouterHost, &config.RouterPort)
		},
	}
	app.Silent = true
	app.Run(os.Args)
}
