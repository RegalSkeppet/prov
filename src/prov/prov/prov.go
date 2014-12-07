package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"prov"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var liveFlag = flag.Bool("live", false, "Actually perform changes.")
var quietFlag = flag.Bool("quiet", false, "Only print summary information.")

func main() {
	log.SetFlags(0)
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("missing file to work on")
	}
	vars, err := parseCommandlineVars(args[1:])
	if err != nil {
		log.Fatal(err.Error())
	}
	err = Provision(args[0], vars, *liveFlag, *quietFlag)
	if err != nil {
		os.Exit(1)
	}
}

func parseCommandlineVars(args []string) (map[interface{}]interface{}, error) {
	vars := map[interface{}]interface{}{}
	for _, arg := range args {
		split := strings.SplitN(arg, "=", 2)
		var v interface{}
		if len(split) > 1 {
			err := yaml.Unmarshal([]byte(split[1]), &v)
			if err != nil {
				return nil, err
			}
		}
		vars[split[0]] = v
	}
	return vars, nil
}

func Provision(filename string, vars map[interface{}]interface{}, live, quiet bool) error {
	if quiet {
		log.SetOutput(ioutil.Discard)
	}
	start := time.Now()
	ok, changed, err := prov.BootstrapFile(filename, vars, live)
	if err != nil {
		log.Println("*** FAILED ***")
		log.SetOutput(os.Stderr)
		log.Print(err.Error())
		return err
	}
	if live {
		log.Println("*** Finished LIVE run ***")
	} else {
		log.Println("*** Finished test run ***")
	}
	log.SetOutput(os.Stderr)
	log.Printf("OK: %d", ok)
	log.Printf("Changes: %d", changed)
	log.Printf("Time: %s", time.Since(start).String())
	return nil
}
