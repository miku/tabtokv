package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"

	"github.com/miku/stardust"
	"github.com/miku/tabtokv"

	// sqlite3 bindings
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	columnSpecVar := flag.String("f", "1,2", "two fields indices, first will be key, second the value")
	version := flag.Bool("v", false, "prints current program version")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	outFile := flag.String("o", "", "output filename")

	var PrintUsage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] MARCFILE\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *version {
		fmt.Println(tabtokv.Version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		PrintUsage()
		os.Exit(1)
	}

	columnSpec, err := stardust.ParseColumnSpec(*columnSpecVar)
	if err != nil {
		log.Fatalln(err)
	}

	filename := flag.Args()[0]
	fi, err := os.Open(filename)

	if err != nil {
		log.Fatalln(err)
	}

	// channel of records
	records := stardust.RecordGeneratorFile(fi, columnSpec)

	// provide ad-hoc temporary filename if none given
	var output string
	if *outFile == "" {
		tmp, err := ioutil.TempFile("", "ntto-")
		output = tmp.Name()
		log.Printf("No explicit [-o]utput given, writing to %s\n", output)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		output = *outFile
	}

	db, err := sql.Open("sqlite3", output)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	init := `CREATE TABLE IF NOT EXISTS store (key text, value text)`
	_, err = db.Exec(init)
	if err != nil {
		log.Fatalf("%q: %s\n", err, init)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	stmt, err := tx.Prepare("INSERT INTO store VALUES (?, ?)")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()

	for record := range records {
		_, err = stmt.Exec(record.Left(), record.Right())
		if err != nil {
			log.Fatalln(err)
		}
	}

	_, err = tx.Exec("CREATE INDEX idx_store_key ON store (key)")
	if err != nil {
		log.Fatalln(err)
	}
	tx.Commit()
}
