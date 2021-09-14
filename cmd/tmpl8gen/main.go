package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/olekukonko/tablewriter"
	copy2 "github.com/otiai10/copy"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	emptyStr         = ""
	specialChars     = ` !@#$%^&*(){}[]|?><,"';:-=~'`
	prompt           = "tmpl8gen> "
	flagTemplatePath = "tmp_path"
	flagMapPath      = "map_path"
	flagOutputDir    = "out_dir"
	flagBypassConf   = "bc"
)

var (
	templateSrcPath     *string
	templateMappingPath *string
	outputPath          *string
	bypassConf          *bool
	subMappings         map[string]string

	// internal vars
	guided bool
	sinRdr *bufio.Reader
)

func main() {
	printBanner()
	sinRdr = bufio.NewReader(os.Stdin)
	setFlags()

	if err := validateFlags(); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		fmt.Println()
		flag.Usage()
		exit()
	}

	// confirm choices
	if !*bypassConf {
		confirmed := confirmEntries()
		if !confirmed {
			fmt.Println("exiting since entries were not confirmed")
			exit()
		}
	}

	if err := loadSubMappings(); err != nil {
		fmt.Println("error while trying to load mappings file:", err)
		exit()
	}

	// create dir
	if err := createDir(); err != nil {
		panic(err)
	}

	// copy template to location
	copyTemplate()

	// generate template
	scaffold()
}

func setFlags() {
	templateSrcPath = flag.String(flagTemplatePath, emptyStr, "path to template directory")
	templateMappingPath = flag.String(flagMapPath, emptyStr, "path to template map file")
	outputPath = flag.String(flagOutputDir, emptyStr, "path for generated output")
	bypassConf = flag.Bool(flagBypassConf, false, "bypass confirmation prompts")

	flag.Parse()
}

func validateFlags() error {
	if flag.NFlag() == 0 {
		return fmt.Errorf("config flags missing")
	}

	*templateSrcPath = strings.TrimSpace(*templateSrcPath)
	*templateMappingPath = strings.TrimSpace(*templateMappingPath)
	*outputPath = strings.TrimSpace(*outputPath)

	if err := validateInput(flagTemplatePath, templateSrcPath); err != nil {
		return err
	}

	if err := validateInput(flagMapPath, templateMappingPath); err != nil {
		return err
	}

	if err := validateInput(flagOutputDir, outputPath); err != nil {
		return err
	}

	return nil
}

func confirmEntries() bool {
	if dir, err := os.Getwd(); err == nil {
		fmt.Printf("Current working dir: %s\n", dir)
	}

	fmt.Println("Please confirm your entries:")

	data := [][]string{
		{"Template Source Dir", *templateSrcPath},
		{"Output Dir", *outputPath},
		{"Mapping Source File", *templateMappingPath},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(true)
	table.AppendBulk(data)
	table.Render()

	fmt.Println("Are these values correct? (y/n)")

	var conf bool
	var cstr string
	var err error

	for len(cstr) == 0 {
		cstr = getInput()

		if strings.EqualFold(cstr, "y") {
			conf = true
			break
		}

		if strings.EqualFold(cstr, "n") {
			conf = false
			break
		}

		conf, err = strconv.ParseBool(cstr)

		if err != nil {
			continue
		}
	}

	return conf
}

func loadSubMappings() error {
	f, err := ioutil.ReadFile(*templateMappingPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(f, &subMappings)
}

func getInput() string {
	fmt.Print(prompt)
	in, _, err := sinRdr.ReadLine()

	if err != nil {
		fmt.Printf("input error: %s\n", err)
		exit()
	}

	sin := string(in)
	sin = strings.TrimSpace(sin)

	if strings.EqualFold(sin, emptyStr) {
		fmt.Println("error:", "value can not be empty")
		return getInput()
	}

	if strings.EqualFold(sin, "q") {
		exit()
		return emptyStr
	}

	fmt.Println()
	return string(in)
}

func createDir() error {
	if _, err := os.Stat(*outputPath); err != nil {
		// create dir
		if err := os.MkdirAll(*outputPath, 0755); err != nil {
			return err
		}
	}
	return nil
}

func copyTemplate() {
	if err := copy2.Copy(*templateSrcPath, *outputPath); err != nil {
		panic(err)
	}
}

func validateInput(key string, val *string) error {
	if strings.ContainsAny(*val, specialChars) {
		return fmt.Errorf("value entered for %s contains invalid chars: %s", key, *val)
	}

	if strings.EqualFold(*val, emptyStr) {
		return fmt.Errorf("value for %s can not be empty", key)
	}

	return nil
}

func exit() {
	os.Exit(0)
}

func scaffold() {
	logData := [][]string{}

	filepath.Walk(*outputPath, func(path string, fi os.FileInfo, pErr error) error {
		fp := path

		if !fi.IsDir() {
			tmpFile, e2 := os.Open(fp)
			if e2 != nil {
				panic(e2)
			}

			scnr := bufio.NewScanner(tmpFile)
			var totalReps int

			changeBuffer := []string{}

			for scnr.Scan() {
				ln := scnr.Text()
				totalReps += findReplaceMatches(&ln)
				changeBuffer = append(changeBuffer, ln)
			}

			tmpFile.Close()

			// first clear contents of file
			terr := os.Truncate(fp, 0)
			if terr != nil {
				panic(terr)
			}

			tmpFile, e2 = os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if e2 != nil {
				panic(e2)
			}

			nln := strings.Join(changeBuffer, "\n")
			_, err := fmt.Fprint(tmpFile, nln)
			if err != nil {
				panic(err)
			}

			logData = append(logData, []string{fp, fmt.Sprintf("%d", totalReps)})
		}

		return nil
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"File Processed", "Total Replacements"})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(true)
	table.AppendBulk(logData)
	table.Render()
}

// returns total replacements
func findReplaceMatches(line *string) int {
	var r int

	for k, v := range subMappings {
		kp := getKeyPattern(k)

		if strings.Contains(*line, kp){
			*line = strings.ReplaceAll(*line, kp, v)
			r++
		}
	}

	return r
}

func getKeyPattern(key string) string {
	return fmt.Sprintf("{%s}", key)
}