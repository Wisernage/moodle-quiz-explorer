package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"timo/internal/api"
	"timo/internal/db"
	"timo/internal/importer"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: timo <command> [options]")
		fmt.Fprintln(os.Stderr, "  timo import <path.mbz>       Import a Moodle course backup (ZIP or tar.gz .mbz; replaces existing quiz data)")
		fmt.Fprintln(os.Stderr, "  timo import-dir <path>       Import an already-extracted backup directory")
		fmt.Fprintln(os.Stderr, "  timo serve [-addr :8080]     Run HTTP UI and API")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "import":
		cmdImport(os.Args[2:])
	case "import-dir":
		cmdImportDir(os.Args[2:])
	case "serve":
		cmdServe(os.Args[2:])
	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}
}

func cmdImport(args []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	dsn := fs.String("dsn", "file:timo.db?cache=shared&_fk=1", "sqlite dsn")
	if err := fs.Parse(args); err != nil {
		log.Fatal(err)
	}
	rest := fs.Args()
	if len(rest) < 1 {
		log.Fatal("usage: timo import [-dsn file:timo.db?...] <path.mbz>")
	}
	path := rest[0]

	if err := db.EnsureDir(*dsn); err != nil {
		log.Fatal(err)
	}
	gdb, err := db.Open(*dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := importer.ImportMBZ(gdb, path); err != nil {
		log.Fatal(err)
	}
	log.Println("import ok:", path)
}

func cmdImportDir(args []string) {
	fs := flag.NewFlagSet("import-dir", flag.ExitOnError)
	dsn := fs.String("dsn", "file:timo.db?cache=shared&_fk=1", "sqlite dsn")
	if err := fs.Parse(args); err != nil {
		log.Fatal(err)
	}
	rest := fs.Args()
	if len(rest) < 1 {
		log.Fatal("usage: timo import-dir [-dsn file:...] <extracted-backup-directory>")
	}
	path := rest[0]

	if err := db.EnsureDir(*dsn); err != nil {
		log.Fatal(err)
	}
	gdb, err := db.Open(*dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := importer.ImportBackupDir(gdb, path); err != nil {
		log.Fatal(err)
	}
	log.Println("import ok:", path)
}

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "listen address")
	dsn := fs.String("dsn", "file:timo.db?cache=shared&_fk=1", "sqlite dsn")
	_ = fs.Parse(args)

	if err := db.EnsureDir(*dsn); err != nil {
		log.Fatal(err)
	}
	gdb, err := db.Open(*dsn)
	if err != nil {
		log.Fatal(err)
	}
	srv := &api.Server{DB: gdb}
	mux := http.NewServeMux()
	srv.Register(mux)
	log.Printf("listening on http://localhost%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
