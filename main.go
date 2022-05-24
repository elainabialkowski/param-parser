package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type field struct {
	name string
	t    string
}

func main() {
	path := os.Args[1]
	if valid := fs.ValidPath(path); !valid {
		log.Fatalf("Path not valid: %s\n", path)
	}

	dir := os.DirFS(path)
	entries, err := fs.Glob(dir, "*.go")
	if err != nil {
		log.Fatalf("Couldn't read directory: %s\n", err.Error())
	}

	for _, v := range entries {
		f, err := os.OpenFile(fmt.Sprintf("%s/%s", path, v), os.O_RDWR, fs.FileMode(os.O_RDWR))
		if err != nil {
			log.Printf("Error while opening file: %s\n", err.Error())
		}
		defer f.Close()

		lines := make([]string, 0)
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
			log.Printf("%s\n", scanner.Text())
		}

		params := make([]field, 0)
		name := ""
		for _, v := range lines {
			isField, err := regexp.MatchString("\\s*[a-zA-Z0-9]+\\s*(string|uint64)", v)
			if err != nil {
				log.Printf("Error while parsing line: %s\n", err.Error())
			}

			if isField {
				param := strings.Fields(v)
				params = append(params, field{
					name: param[0],
					t:    param[1],
				})
			}

			isName, err := regexp.MatchString("(type)\\s*[a-zA-Z0-9]+\\s*(struct)", v)
			if err != nil {
				log.Printf("Error while parsing line: %s\n", err.Error())
			}

			if isName {
				name = strings.Fields(v)[1]
			}

		}

		log.Printf("Type: %s\n", name)
		for _, v := range params {
			log.Printf("%s %s\n", v.name, v.t)
		}

		_, err = f.WriteString(fmt.Sprintf(`
			func (%s) populate(params string) (%s, error) {
				var err error
				split := strings.Fields(params)
				populated := %s{}
				%s
				return populated, nil
			}
		`,
			name,
			name,
			name,
			Parse(params, strings.ToLower(string(name[0])), name),
		))
		if err != nil {
			log.Printf("Error while writing file: %s\n", err.Error())
		}

		err = exec.Command("gopls", "fix", "-a", "-w", fmt.Sprintf("%s/%s", path, v)).Run()
		if err != nil {
			log.Printf("Error while formatting: %s\n", err.Error())
		}

		err = exec.Command("gopls", "format", "-w", fmt.Sprintf("%s/%s", path, v)).Run()
		if err != nil {
			log.Printf("Error while formatting: %s\n", err.Error())
		}

		err = exec.Command("gopls", "imports", "-w", fmt.Sprintf("%s/%s", path, v)).Run()
		if err != nil {
			log.Printf("Error while formatting: %s\n", err.Error())
		}

	}

}

func Parse(f []field, name, t string) string {
	result := strings.Builder{}
	for k, v := range f {
		switch v.t {
		case "string":
			result.WriteString(fmt.Sprintf(`
				populated.%s = split[%d]
			`, v.name, k))
			break
		case "uint64":
			result.WriteString(fmt.Sprintf(`
				populated.%s, err = strconv.ParseUint(split[%d], 10, 64)
				if err != nil {
					return %s{}, err
				}
			`, v.name, k, t))
			break
		}
	}
	return result.String()
}
