// Copyright 2014 by Leipzig University Library, http://ub.uni-leipzig.de
//                   The Finc Authors, http://finc.info
//                   Martin Czygan, <martin.czygan@uni-leipzig.de>
//
// This file is part of some open source application.
//
// Some open source application is free software: you can redistribute
// it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation, either
// version 3 of the License, or (at your option) any later version.
//
// Some open source application is distributed in the hope that it will
// be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
// of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
//
// @license GPL-3.0+ <http://spdx.org/licenses/GPL-3.0+>
//
//
// tabtokv is a command line too to turn two CSV columns into a sqlite key-value store.
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
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] TAB-SEPARATED-FILE\n", os.Args[0])
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
