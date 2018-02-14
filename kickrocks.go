package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"
	"os"
	"path/filepath"
)

func main() {
	// Get the working dir
	pwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	// Allow working directory override; default to parent working directory
	pwdInput := flag.String("dir", pwd, "a path to rotate")

	// Flags to skip dotfiles and sub-dirs
	skipDotenvFlag := flag.Bool("skipdotfiles", false, "a flag to prevent dotfiles (.env) from being rotated")
	skipDirsFlag := flag.Bool("skipsubdirs", false, "a flag to prevent sub-directories from being rotated")

	// Flag to show output
	verboseFlag := flag.Bool("verbose", false, "a flag to show output")

	flag.Parse()

	pwd = *pwdInput
	var skipDotenvs bool = *skipDotenvFlag
	var skipDirs bool = *skipDirsFlag
	var verbose bool = *verboseFlag

	// Define rotate paths
	rotatePaths := make(map[string]string)
	rotatePaths["last_month"] = filepath.Join(pwd, "last_month")
	rotatePaths["last_week"] = filepath.Join(pwd, "last_week")

	rotate := func(src string, dest string, after int64) {
		// Get files in src directory
		files, err := ioutil.ReadDir(src)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			var name string = f.Name()
			if name[0] == '.' && skipDotenvs {
				continue
			} else if f.IsDir() && skipDirs {
				continue
			}

			// Skip protected directories
			var fp bool = false
			for name, path := range rotatePaths {
				// Only skip if in parent working directory
				if (name == f.Name() && pwd == src) {
					if (verbose) {
						fmt.Println(path, "is a protected directory; Skipping")
					}
					fp = true
					break
				}
			}
			if fp {
				continue
			}

			// Rotate top-level resources
			var hours = int64(time.Since(f.ModTime()).Hours())
			var days = int64(hours / 24)

			if days >= after {
				if (verbose) {
					fmt.Println(filepath.Join(src, f.Name()), "moved to", filepath.Join(dest, f.Name()))
				}
				os.Rename(filepath.Join(src, f.Name()), filepath.Join(dest, f.Name()))
			}
			switch {
			case days >= 30:
				if (verbose) {
					fmt.Println(f.Name(), "to be moved to last_month dir");
				}
				break
			case days >= 7:
				if (verbose) {
					fmt.Println(f.Name(), "to be moved to last_week dir");
				}
				break;
			}
		}
		return
	}

	// Make sure that rotate paths actual exist
	for _, rotatePath := range rotatePaths {
		if _, err := os.Stat(rotatePath); os.IsNotExist(err) {
			os.Mkdir(rotatePath, 0775)
		}
	}

	// Rotate files in last_week that are > 14 days old to the last_month directory
	rotate(rotatePaths["last_week"], rotatePaths["last_month"], 14);
	rotate(pwd, rotatePaths["last_week"], 7);
}
