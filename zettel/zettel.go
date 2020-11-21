package zettel

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kraem/zhuyi-go/pkg/fs"
	"github.com/kraem/zhuyi-go/pkg/log"
)

const yamlFmDelim = "---"
const yamlFmTitleField = "title: "
const yamlFmDateField = "date: "

const mdExtension = ".md"

const index = "index"

const timeFormatFile = "060102-1504"
const timeFormatFm = "2006-01-02 15:04"

const abc = "abcdefghijklmnopqrstuvwxyz"

// regex
// x matches x exactly once
// x{a,b} matches x between a and b times
// x{a,} matches x at least a times
// x{,b} matches x up to (a maximum of) b times
// x* matches x zero or more times (same as x{0,})
// x+ matches x one or more times (same as x{1,})
// x? matches x zero or one time (same as x{0,1})
var typeExtractor = regexp.MustCompile("(.*): (.*)")
var linkExtractor = regexp.MustCompile(`\[([^\[]*)\](\((.*)\))`)

type Zettel struct {
	// TODO
	// remove these json tags
	// and convert to another payload struct
	Date  string   `json:"date"`
	Title string   `json:"title"`
	File  string   `json:"file"`
	Links []string `json:"links"`
}

func extractMarkdownLinks(path string) (links []string, err error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {

		tokens := linkExtractor.FindStringSubmatch(scanner.Text())

		if len(tokens) > 1 {
			links = append(links, tokens[3])
		}
	}

	return
}

func extractFrontMatterFields(path string) (fields map[string]string, err error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	fields = make(map[string]string)
	separatorOccourcencesRemaining := 2
	for scanner.Scan() && separatorOccourcencesRemaining > 0 {
		if scanner.Text() == yamlFmDelim {
			separatorOccourcencesRemaining--
		}

		tokens := typeExtractor.FindStringSubmatch(scanner.Text())

		if len(tokens) > 1 {
			fields[tokens[1]] = tokens[2]
		}
	}

	return fields, nil
}

func (c *Config) UnlinkedNodes() ([]Zettel, error) {
	fim, err := c.mapFileToInt(true)
	if err != nil {
		return nil, err
	}

	indexNode := fim.fileToInt[c.ZettelPath+index+mdExtension]
	walked := make(map[int]bool, len(fim.fileToInt))
	c.walkGraph(indexNode, fim, walked)

	return zettelsFromWalked(walked, fim)
}

// walkGraph walks the graph depth first
func (c *Config) walkGraph(n int, g *fileIntMap, walked map[int]bool) {
	walked[n] = true
	fn := g.intToFile[n]
	if !strings.HasSuffix(fn, mdExtension) {
		return
	}
	links, err := extractMarkdownLinks(fn)
	if err != nil {
		log.LogError(err)
		return
	}
	for _, l := range links {
		lp := c.ZettelPath + l
		li := g.fileToInt[lp]
		if !walked[li] {
			c.walkGraph(li, g, walked)
		}
	}
}

func zettelsFromWalked(walked map[int]bool, fim *fileIntMap) ([]Zettel, error) {
	unlinked := make([]Zettel, 0)
	for k, v := range fim.intToFile {
		if !walked[k] {
			fmFields, err := extractFrontMatterFields(v)
			if err != nil {
				log.LogError(err)
				continue
			}
			_, fileName := path.Split(v)
			z := Zettel{
				Title: fmFields["title"],
				File:  fileName,
				Date:  fmFields["date"],
				// TODO
				// operate on zettels like objects will make
				// this come by itself
				//Links: links,
			}
			unlinked = append(unlinked, z)
		}
	}
	return unlinked, nil
}

// TODO
// this is broken..
func (c *Config) FindIsolatedVertices() ([]Zettel, error) {

	// unfortunately we need to traverse the dir
	// two times since we need the number of files,
	// and their mappings to ints (and back to file names)
	// for the adjacentmatrix first
	fim, err := c.mapFileToInt(false)
	if err != nil {
		return nil, err
	}

	adjMatrix, fileZsMap, err := c.buildAdjacencyMatrix(*fim)
	if err != nil {
		return nil, err
	}

	vs := findIsolatedVertices(adjMatrix)
	filteredZs := make([]Zettel, 0)
	for _, i := range vs {
		fn := fim.intToFile[i]
		filteredZs = append(filteredZs, fileZsMap[fn])
	}

	return filteredZs, nil
}

// creating a representation of a directed graph with an adjacency matrix
//
// example of graph:
//
//	    a -> b	  ( a links to b )
// 	    b		  ( b links to nothing )
//
// result:
//
// - is '[' and ']' turned 90 degrees
//
//	    - [ a b ]
// 	    a   0 1
// 	    b   0 0
// 	    -
// also:
// returning the zettels mapped to filename here so we don't want to iterate once more
// to create them. we're already extracting fm here anyway..
func (c *Config) buildAdjacencyMatrix(fim fileIntMap) ([][]int, map[string]Zettel, error) {
	path := c.ZettelPath

	adjMatrix := make([][]int, len(fim.fileToInt))
	for i := range adjMatrix {
		adjMatrix[i] = make([]int, len(fim.fileToInt))
	}

	zs := make(map[string]Zettel, 0)

	// could be implemented with filepath.Walk(path, func(path string, info os.Fileinfo, errerror) error { //do stuff })
	// but let's keep it simple as we don't use subdirs
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, nil, err
	}
	for _, f := range files {
		fileName := f.Name()

		if !strings.HasSuffix(fileName, mdExtension) {
			continue
		}

		fullPath := filepath.Join(path, fileName)

		fmFields, err := extractFrontMatterFields(fullPath)
		if err != nil {
			log.LogError(err)
			continue
		}

		links, err := extractMarkdownLinks(fullPath)
		if err != nil {
			log.LogError(err)
			continue
		}

		z := Zettel{
			Title: fmFields["title"],
			File:  fileName,
			Date:  fmFields["date"],
			Links: links,
		}

		zs[fileName] = z

		for _, linkedFile := range links {
			origFileInt := fim.fileToInt[fileName]
			linkedFileInt := fim.fileToInt[linkedFile]
			adjMatrix[origFileInt][linkedFileInt] = 1
		}
	}

	return adjMatrix, zs, nil
}

// TODO
// write test
func findIsolatedVertices(adjMatrix [][]int) []int {
	isolatedVs := make([]int, 0)
	for i := range adjMatrix {
		if sumInts(adjMatrix[i]) < 1 {
			colSum := 0
			for j := range adjMatrix[i] {
				colSum += adjMatrix[j][i]
			}
			if colSum == 0 {
				isolatedVs = append(isolatedVs, i)
			}
		}
	}
	return isolatedVs
}

func sumInts(slice []int) int {
	sum := 0
	for _, i := range slice {
		sum += i
	}
	return sum
}

type fileIntMap struct {
	fileToInt map[string]int
	intToFile map[int]string
}

func newFileIntMap() fileIntMap {
	return fileIntMap{
		fileToInt: make(map[string]int, 0),
		intToFile: make(map[int]string, 0),
	}
}

func (c *Config) mapFileToInt(fullPath bool) (*fileIntMap, error) {
	path := c.ZettelPath

	nrVertices := 0

	fim := newFileIntMap()

	// could be implemented with filepath.Walk(path, func(path string, info os.Fileinfo, errerror) error { //do stuff })
	// but let's keep it simple as we don't use subdirs
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		fileName := f.Name()

		if f.IsDir() {
			continue
		}

		if !strings.HasSuffix(fileName, mdExtension) {
			continue
		}

		if fullPath {
			fileName = filepath.Join(path, fileName)
		}

		fim.fileToInt[fileName] = nrVertices
		fim.intToFile[nrVertices] = fileName

		nrVertices += 1
	}

	return &fim, nil
}

func (c *Config) CreateZettel(title, body string) (fileName string, err error) {
	timeNow := time.Now()
	zettelDateFm := timeNow.Format(timeFormatFm)
	zettelFileName := timeNow.Format(timeFormatFile)
	zettelFilePath := c.ZettelPath + zettelFileName + mdExtension

	exist, err := fs.PathExists(zettelFilePath)
	if err != nil {
		log.LogError(err)
		os.Exit(1)
	}

	if exist {
		for i := range abc {
			if i == 0 {
				zettelFileName = zettelFileName + string(abc[i])
				zettelFilePath = c.ZettelPath + zettelFileName + mdExtension
			} else if i == (len(abc) - 1) {
				err := fmt.Errorf("too many files created in one minute")
				log.LogError(err)
				return "", err
			} else {
				runes := []rune(zettelFileName)
				runes[len(zettelFileName)-1] = rune(abc[i])
				zettelFileName = string(runes)
				zettelFilePath = c.ZettelPath + zettelFileName + mdExtension
			}
			exist, err = fs.PathExists(zettelFilePath)
			if err != nil {
				log.LogError(err)
				return "", err
			}
			if exist {
				continue
			}
			break
		}

	}

	f, err := os.Create(zettelFilePath)
	if err != nil {
		log.LogError(err)
		return "", err
	}
	defer f.Close()

	f.Write([]byte(yamlFmDelim + "\n"))
	f.Write([]byte(yamlFmTitleField + title + "\n"))
	f.Write([]byte(yamlFmDateField + zettelDateFm + "\n"))
	f.Write([]byte(yamlFmDelim + "\n\n"))
	f.Write([]byte(body + "\n"))

	zettelFileName = zettelFileName + mdExtension

	return zettelFileName, nil
}

// TODO
// D3 area
// 1. extract to own module?
// 2. test with fe

type D3jsGraph struct {
	Nodes []Node `json:"nodes,omitempty"`
	Links []Link `json:"links,omitempty"`
}

type Node struct {
	Id     string `json:"id"`
	Radius string `json:"radius"`
}

type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  string `json:"value"`
}

func (c *Config) CreateD3jsGraph() (*D3jsGraph, error) {
	d3AdjList, err := buildD3AdjacancyList(c.ZettelPath)
	if err != nil {
		return nil, err
	}

	var g D3jsGraph

	for _, z := range d3AdjList {

		radius := strconv.Itoa(len(z.Links) + 2)

		outputNode := Node{
			Id:     z.Title,
			Radius: radius,
		}

		g.Nodes = append(g.Nodes, outputNode)

		for i := range z.Links {
			link := z.Links[i]
			_, ok := d3AdjList[link]
			switch ok {
			// TODO
			// can't remember why we differantiate between these
			case true:
				outputLink := Link{
					Source: z.Title,
					Target: d3AdjList[link].Title,
					Value:  "2",
				}
				g.Links = append(g.Links, outputLink)
			case false:
				outputNode := Node{
					Id:     link,
					Radius: "2",
				}

				g.Nodes = append(g.Nodes, outputNode)
				outputLink := Link{
					Source: z.Title,
					Target: outputNode.Id,
					Value:  "2",
				}
				g.Links = append(g.Links, outputLink)
			}
		}
	}

	return &g, nil
}

// returns hash map of file names
// each filename contains a zettel object
// which in turn includes all of its
// filenames/http-links it links to
func buildD3AdjacancyList(path string) (map[string]Zettel, error) {
	fileToZettelMap := make(map[string]Zettel)

	// could be implemented with filepath.Walk(path, func(path string, info os.Fileinfo, errerror) error { //do stuff })
	// but let's keep it simple as we don't use subdirs
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		fileName := f.Name()

		if !strings.HasSuffix(fileName, mdExtension) {
			continue
		}

		fullPath := filepath.Join(path, fileName)

		fmFields, err := extractFrontMatterFields(fullPath)
		if err != nil {
			log.LogError(err)
			continue
		}

		// TODO
		// this is probably not we want..
		if _, exists := fmFields["title"]; !exists {
			continue
		}

		links, err := extractMarkdownLinks(fullPath)
		if err != nil {
			log.LogError(err)
			continue
		}

		zettel := Zettel{
			Title: fmFields["title"],
			File:  fileName,
			Date:  fmFields["date"],
			Links: links,
		}

		fileToZettelMap[fileName] = zettel
	}

	return fileToZettelMap, nil
}
