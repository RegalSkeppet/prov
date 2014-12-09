package main

import (
	"flag"
	"log"
	"os"
	"prov"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var liveFlag = flag.Bool("live", false, "Actually perform changes.")

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
	err = Provision(args[0], vars, *liveFlag)
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

func Provision(filename string, vars map[interface{}]interface{}, live bool) error {
	start := time.Now()
	ok, changed, skipped, err := prov.RunFile(filename, vars, live)
	if err != nil {
		log.Println("*** FAILED ***")
		log.Print(err.Error())
		return err
	}
	if live {
		log.Println("*** Finished LIVE run ***")
	} else {
		log.Println("*** Finished test run ***")
	}
	log.Printf("time: %s\n", time.Since(start).String())
	log.Printf("skipped: %d\n", skipped)
	log.Printf("ok: %d\n", ok)
	log.Printf("CHANGED: %d\n", changed)
	return nil
}
