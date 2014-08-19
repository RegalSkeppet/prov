package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"prov"
	"time"
)

var varsFlag prov.Vars = prov.Vars{}
var runFlag = flag.Bool("run", false, "Actually perform changes.")
var quietFlag = flag.Bool("quiet", false, "Only print summary information.")

func init() {
	flag.Var(&varsFlag, "vars", "YAML map to use as variables. Overrides existing variables.")
}

func main() {
	log.SetFlags(0)
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("missing file to work on")
	}
	err := Provision(args[0], varsFlag, *runFlag, *quietFlag)
	if err != nil {
		os.Exit(1)
	}
}

func Provision(filename string, vars prov.Vars, run, quiet bool) error {
	if quiet {
		log.SetOutput(ioutil.Discard)
	}
	if *runFlag {
		log.Println("*** Starting LIVE run ***")
	} else {
		log.Println("*** Starting test run ***")
	}
	log.Println()
	start := time.Now()
	ok, changed, err := prov.BootstrapFile(filename, vars, run)
	if err != nil {
		log.Println("*** FAILED ***")
		log.SetOutput(os.Stderr)
		log.Print(err.Error())
		return err
	}
	if *runFlag {
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
