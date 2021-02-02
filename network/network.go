package network

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
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

type Node struct {
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

func (c *Config) UnlinkedNodes() ([]Node, error) {
	fim, err := c.mapFileToInt(true)
	if err != nil {
		return nil, err
	}

	indexNode := fim.fileToInt[c.NetworkPath+index+mdExtension]
	walked := make(map[int]bool, len(fim.fileToInt))
	c.walkGraph(indexNode, fim, walked)

	ns, err := nodesFromWalked(walked, fim)
	if err != nil {
		return nil, err
	}
	sortNodesDate(ns)
	return ns, nil
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
		lp := c.NetworkPath + l
		li := g.fileToInt[lp]
		if !walked[li] {
			c.walkGraph(li, g, walked)
		}
	}
}

func nodesFromWalked(walked map[int]bool, fim *fileIntMap) ([]Node, error) {
	unlinked := make([]Node, 0)
	for k, v := range fim.intToFile {
		if !walked[k] {
			fmFields, err := extractFrontMatterFields(v)
			if err != nil {
				log.LogError(err)
				continue
			}
			_, fileName := path.Split(v)
			n := Node{
				Title: fmFields["title"],
				File:  fileName,
				Date:  fmFields["date"],
				// TODO
				// operate on nodes like objects will make
				// this come by itself
				//Links: links,
			}
			unlinked = append(unlinked, n)
		}
	}
	return unlinked, nil
}

func sortNodesDate(ns []Node) {
	// file names are the dates..
	sort.Slice(ns, func(i, j int) bool {
		return ns[i].File > ns[j].File
	})
}

// TODO
// this is broken..
func (c *Config) FindIsolatedVertices() ([]Node, error) {

	// unfortunately we need to traverse the dir
	// two times since we need the number of files,
	// and their mappings to ints (and back to file names)
	// for the adjacentmatrix first
	fim, err := c.mapFileToInt(false)
	if err != nil {
		return nil, err
	}

	adjMatrix, fileNsMap, err := c.buildAdjacencyMatrix(*fim)
	if err != nil {
		return nil, err
	}

	vs := findIsolatedVertices(adjMatrix)
	filteredNs := make([]Node, 0)
	for _, i := range vs {
		fn := fim.intToFile[i]
		filteredNs = append(filteredNs, fileNsMap[fn])
	}

	return filteredNs, nil
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
// returning the nodes mapped to filename here so we don't want to iterate once more
// to create them. we're already extracting fm here anyway..
func (c *Config) buildAdjacencyMatrix(fim fileIntMap) ([][]int, map[string]Node, error) {
	path := c.NetworkPath

	adjMatrix := make([][]int, len(fim.fileToInt))
	for i := range adjMatrix {
		adjMatrix[i] = make([]int, len(fim.fileToInt))
	}

	ns := make(map[string]Node, 0)

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

		n := Node{
			Title: fmFields["title"],
			File:  fileName,
			Date:  fmFields["date"],
			Links: links,
		}

		ns[fileName] = n

		for _, linkedFile := range links {
			origFileInt := fim.fileToInt[fileName]
			linkedFileInt := fim.fileToInt[linkedFile]
			adjMatrix[origFileInt][linkedFileInt] = 1
		}
	}

	return adjMatrix, ns, nil
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
	path := c.NetworkPath

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

func (c *Config) DelNode(filename string) error {
	fp := filepath.Join(c.NetworkPath, filename)
	exist, err := fs.PathExists(fp)
	if err != nil {
		return err
	}
	// TODO
	// this might not be needed
	// as os.Remove calls unlink..
	if !exist {
		err := fmt.Errorf("file doesn't exist: %v", fp)
		return err
	}
	err = os.Remove(fp)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) CreateNode(title, body string) (fileName string, err error) {
	timeNow := time.Now()
	nodeDateFm := timeNow.Format(timeFormatFm)
	nodeFileName := timeNow.Format(timeFormatFile)
	nodeFilePath := c.NetworkPath + nodeFileName + mdExtension

	exist, err := fs.PathExists(nodeFilePath)
	if err != nil {
		err := fmt.Errorf("file already exists")
		return "", err
	}

	if exist {
		for i := range abc {
			if i == 0 {
				nodeFileName = nodeFileName + string(abc[i])
				nodeFilePath = c.NetworkPath + nodeFileName + mdExtension
			} else if i == (len(abc) - 1) {
				err := fmt.Errorf("too many files created in one minute")
				log.LogError(err)
				return "", err
			} else {
				runes := []rune(nodeFileName)
				runes[len(nodeFileName)-1] = rune(abc[i])
				nodeFileName = string(runes)
				nodeFilePath = c.NetworkPath + nodeFileName + mdExtension
			}
			exist, err = fs.PathExists(nodeFilePath)
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

	f, err := os.Create(nodeFilePath)
	if err != nil {
		log.LogError(err)
		return "", err
	}
	defer f.Close()

	f.Write([]byte(yamlFmDelim + "\n"))
	f.Write([]byte(yamlFmTitleField + title + "\n"))
	f.Write([]byte(yamlFmDateField + nodeDateFm + "\n"))
	f.Write([]byte(yamlFmDelim + "\n\n"))
	f.Write([]byte(body + "\n"))

	nodeFileName = nodeFileName + mdExtension

	return nodeFileName, nil
}

// TODO
// D3 area
// 1. extract to own module?
// 2. test with fe

type D3jsGraph struct {
	Nodes []D3Node `json:"nodes,omitempty"`
	Links []Link   `json:"links,omitempty"`
}

type D3Node struct {
	Id     string `json:"id"`
	Radius string `json:"radius"`
}

type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  string `json:"value"`
}

func (c *Config) CreateD3jsGraph() (*D3jsGraph, error) {
	filenameToNode, err := linksPerFilename(c.NetworkPath)
	if err != nil {
		return nil, err
	}

	var g D3jsGraph

	for _, n := range filenameToNode {

		radius := strconv.Itoa(len(n.Links) + 2)

		outputNode := D3Node{
			Id:     n.Title,
			Radius: radius,
		}

		g.Nodes = append(g.Nodes, outputNode)

		for i := range n.Links {
			link := n.Links[i]
			_, ok := filenameToNode[link]
			switch ok {
			// TODO
			// fix bug where we create multiple nodes for http-links.
			// they are created since we don't verify we have created
			// them before..
			case true:
				outputLink := Link{
					Source: n.Title,
					Target: filenameToNode[link].Title,
					Value:  "2",
				}
				g.Links = append(g.Links, outputLink)
			case false:
				outputNode := D3Node{
					Id:     link,
					Radius: "2",
				}

				g.Nodes = append(g.Nodes, outputNode)
				outputLink := Link{
					Source: n.Title,
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
// each filename contains a node object
// which in turn includes all of its
// filenames/http-links it links to
func linksPerFilename(path string) (map[string]Node, error) {
	fileToNodeMap := make(map[string]Node)

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
		//if _, exists := fmFields["title"]; !exists {
		//	continue
		//}

		links, err := extractMarkdownLinks(fullPath)
		if err != nil {
			log.LogError(err)
			continue
		}

		n := Node{
			Title: fmFields["title"],
			File:  fileName,
			Date:  fmFields["date"],
			Links: links,
		}

		fileToNodeMap[fileName] = n
	}

	return fileToNodeMap, nil
}
