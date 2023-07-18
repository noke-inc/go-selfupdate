package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/kr/binarydist"
)

var version, genDir string

type current struct {
	Version string
	Path    string
	Sha256  []byte
}

func generateSha256(path string) []byte {
	h := sha256.New()
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
	}
	h.Write(b)
	sum := h.Sum(nil)
	return sum
	//return base64.URLEncoding.EncodeToString(sum)
}

type gzReader struct {
	z, r io.ReadCloser
}

func (g *gzReader) Read(p []byte) (int, error) {
	return g.z.Read(p)
}

func (g *gzReader) Close() error {
	g.z.Close()
	return g.r.Close()
}

func newGzReader(r io.ReadCloser) io.ReadCloser {
	var err error
	g := new(gzReader)
	g.r = r
	g.z, err = gzip.NewReader(r)
	if err != nil {
		panic(err)
	}
	return g
}

func createUpdate(path string, platform string) {
	fmt.Println("createUpdate " + genDir)
	var c []current
	files, err := ioutil.ReadDir(genDir)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("createUpdate")
	for _, file := range files {
		if file.IsDir() {
			filePath := "./" + filepath.Join(genDir, file.Name())
			c = append(c, current{Version: file.Name(), Path: filePath, Sha256: generateSha256(filePath)})
		}
	}
	fmt.Println("createUpdate")
	c = append(c, current{Version: version, Path: path, Sha256: generateSha256(path)})

	fmt.Println("createUpdate")
	b, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		fmt.Println("error:", err)
	}
	err = ioutil.WriteFile(filepath.Join(genDir, platform+".json"), b, 0755)
	if err != nil {
		panic(err)
	}

	fmt.Println("createUpdate")
	os.MkdirAll(filepath.Join(genDir, version), 0755)

	fmt.Println("createUpdate")
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	f, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	w.Write(f)
	w.Close() // You must close this first to flush the bytes to the buffer.
	err = ioutil.WriteFile(filepath.Join(genDir, version, platform+".gz"), buf.Bytes(), 0755)
	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println("createUpdate")
	for _, file := range files {

		fmt.Println("createUpdate range")
		if !file.IsDir() {
			continue
		}
		if file.Name() == version {
			continue
		}

		fmt.Println("createUpdate range")
		os.Mkdir(filepath.Join(genDir, file.Name(), version), 0755)

		fName := filepath.Join(genDir, file.Name(), platform+".gz")
		old, err := os.Open(fName)
		if err != nil {
			// Don't have an old release for this os/arch, continue on
			continue
		}

		fName = filepath.Join(genDir, version, platform+".gz")
		newF, err := os.Open(fName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't open %s: error: %s\n", fName, err)
			os.Exit(1)
		}

		ar := newGzReader(old)
		defer ar.Close()
		br := newGzReader(newF)
		defer br.Close()
		patch := new(bytes.Buffer)
		if err := binarydist.Diff(ar, br, patch); err != nil {
			panic(err)
		}
		ioutil.WriteFile(filepath.Join(genDir, file.Name(), version, platform), patch.Bytes(), 0755)
	}
}

func makeAllUpdate() {
	// files, err := ioutil.ReadDir(genDir)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	fmt.Println("makeAllUpdate")
}

func printUsage() {
	fmt.Println("")
	fmt.Println("Positional arguments:")
	fmt.Println("\tSingle platform: go-selfupdate myapp 1.2")
	fmt.Println("\tCross platform: go-selfupdate /tmp/mybinares/ 1.2")
}

func createBuildDir() {
	os.MkdirAll(genDir, 0755)
}

func main() {
	fmt.Println("main")
	outputDirFlag := flag.String("o", "public", "Output directory for writing updates")

	var defaultPlatform string
	goos := os.Getenv("GOOS")
	goarch := os.Getenv("GOARCH")
	if goos != "" && goarch != "" {
		defaultPlatform = goos + "-" + goarch
	} else {
		defaultPlatform = runtime.GOOS + "-" + runtime.GOARCH
	}
	platformFlag := flag.String("platform", defaultPlatform,
		"Target platform in the form OS-ARCH. Defaults to running os/arch or the combination of the environment variables GOOS and GOARCH if both are set.")

	fmt.Println("main")
	flag.Parse()
	if flag.NArg() < 2 {
		flag.Usage()
		printUsage()
		os.Exit(0)
	}

	fmt.Println("main")
	platform := *platformFlag
	appPath := flag.Arg(0)
	version = flag.Arg(1)
	genDir = *outputDirFlag

	createBuildDir()

	fmt.Println("main")
	// If dir is given create update for each file
	fi, err := os.Stat(appPath)
	if err != nil {
		panic(err)
	}

	fmt.Println("main appPath: %v, platform: %v", appPath, platform)
	if fi.IsDir() {
		fmt.Println("main")
		files, err := ioutil.ReadDir(appPath)
		if err == nil {
			for _, file := range files {
				fmt.Println("main blah")
				createUpdate(filepath.Join(appPath, file.Name()), file.Name())
			}
			fmt.Println("agh")
			os.Exit(0)
		}
	}

	fmt.Println("main appPath: %v, platform: %v", appPath, platform)
	createUpdate(appPath, platform)
	fmt.Println("main")
}
