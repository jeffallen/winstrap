// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func init() {
	altMain = notWindowsMain
}

var buildWindows = flag.Bool("newwin", false, "force a make.bash of windows_386")

func notWindowsMain() {
	build := flag.Bool("build", false, "build winstrap.exe")
	upload := flag.Bool("upload", false, "upload winstrap.exe to code.google.com (implies -build)")
	flag.Parse()
	if *upload {
		*build = true
	}
	if !*build {
		log.Printf("Not running on Windows and no flags specified.")
		flag.PrintDefaults()
		return
	}
	digest := buildWinstrap()
	if *upload {
		f, err := os.Open("winstrap.exe")
		check(err)
		defer f.Close()
		date := time.Now().Format("2006-01-02")
		fileName := fmt.Sprintf("winstrap-%s-%s.exe", date, digest[:7])
		check(Upload(fileName, "winstrap.exe from "+date+": "+digest,
			f))
		log.Printf("uploaded %s", fileName)
	}
}

// buildWinstrap builds a new winstrap.exe and returns its sha1
func buildWinstrap() string {
	log.Printf("building new winstrap.exe")

	goRoot := os.Getenv("GOROOT")
	if goRoot == "" {
		log.Fatal("no GOROOT set")
	}
	pkgWin := filepath.Join(goRoot, "pkg", "windows_386")
	if _, err := os.Stat(pkgWin); err != nil {
		if os.IsNotExist(err) {
			log.Printf("no %s directory, need to build windows cross-compiler", pkgWin)
			*buildWindows = true
		} else {
			log.Fatal(err)
		}
	}
	winEnv := append([]string{"CGO_ENABLED=0", "GOOS=windows", "GOARCH=386"}, os.Environ()...)
	if *buildWindows {
		cmd := exec.Command(filepath.Join(goRoot, "src", "make.bash"))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = winEnv
		cmd.Dir = filepath.Join(goRoot, "src")
		check(cmd.Run())
	}
	cmd := exec.Command("go", "build", "-o", "winstrap.exe")
	cmd.Env = winEnv
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error building winstrap.exe %v: %s", err, out)
	}
	f, err := os.Open("winstrap.exe")
	check(err)
	defer f.Close()
	s1 := sha1.New()
	_, err = io.Copy(s1, f)
	check(err)
	digest := fmt.Sprintf("%x", s1.Sum(nil))
	log.Printf("built new winstrap.exe; %s", digest)
	return digest
}
