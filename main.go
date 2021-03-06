package main

import "fmt"

import "strconv"
import "github.com/codegangsta/cli"
import "time"
import "math/rand"
import "io/ioutil"

import "strings"
import "os"
import "os/exec"

//import "sort"

import "github.com/antonholmquist/jason"

func main() {
	rand.Seed(time.Now().UnixNano())
	app := cli.NewApp()
	app.Name = "alamode"
	app.Usage = "alamode command line interface for mode"
	app.Version = "16"
	app.Commands = []cli.Command{
		{Name: "spaces", ShortName: "sp",
			Usage: "spaces", Action: SpacesAction},
		{Name: "reports", ShortName: "re",
			Usage: "reports", Action: ReportsAction},
		{Name: "queries", ShortName: "q",
			Usage: "queries", Action: QueriesAction},
		{Name: "sql", ShortName: "s",
			Usage: "sql", Action: SqlAction},
		{Name: "run", ShortName: "r",
			Usage: "run", Action: RunAction},
	}

	app.Run(os.Args)
}

func handleThing(thing, meta string, print bool) []*jason.Object {
	v, _ := jason.NewObjectFromBytes([]byte(thing))
	if v == nil {
		return []*jason.Object{}
	}
	e, _ := v.GetObject("_embedded")
	s, _ := e.GetObjectArray(meta)
	//token name
	list := []string{}
	list_names := []string{}
	for i, item := range s {
		stoken, _ := item.GetString("token")
		sname, _ := item.GetString("name")
		list = append(list, stoken)
		list_names = append(list_names, sname)
		if print {
			fmt.Printf("%d. %s %s\n", i+1, stoken, sname)
		}
	}
	SaveList(meta, list)
	SaveList(meta+"_names", list_names)
	return s
}
func conf() map[string]string {
	m := map[string]string{}
	b, _ := ioutil.ReadFile("conf/settings")
	prev := ""
	for i, line := range strings.Split(string(b), "\n") {
		if i%2 == 0 {
			prev = line
		} else {
			m[line] = prev
		}
	}
	return m
}
func SpacesAction(c *cli.Context) {
	istr := c.Args().Get(0)
	if istr != "" {
		i, _ := strconv.Atoi(istr)
		list := ReadList("spaces")
		SaveLast("space", list[i-1])
		return
	}
	spaces := DoVerb("spaces")
	handleThing(spaces, "spaces", true)
}
func ReportsAction(c *cli.Context) {
	istr := c.Args().Get(0)
	if istr != "" {
		i, _ := strconv.Atoi(istr)
		list := ReadList("reports")
		SaveLast("report", list[i-1])
		return
	}

	spaceToken := ReadLast("space")
	reports := DoVerb("spaces/" + spaceToken + "/reports")
	handleThing(reports, "reports", true)
}
func QueriesAction(c *cli.Context) {
	istr := c.Args().Get(0)
	if istr != "" {
		i, _ := strconv.Atoi(istr)
		list := ReadList("queries")
		list_names := ReadList("queries_names")
		SaveLast("query", list[i-1])
		SaveLast("query_name", list_names[i-1])
		return
	}

	token := ReadLast("report")

	queries := DoVerb("reports/" + token + "/queries")
	handleThing(queries, "queries", true)
}
func RunAction(c *cli.Context) {
	m := conf()
	dsi_s := m["data_source_id"]
	dsi, _ := strconv.Atoi(dsi_s)
	rToken := ReadLast("report")
	qToken := ReadLast("query")
	qName := ReadLast("query_name")

	r := DoVerb("reports/" + rToken + "/queries/" + qToken + "/runs")
	cra := handleLinks(r, "query_runs", false)

	sql := ReadSQL(qToken)
	query := map[string]interface{}{"create_query_run": true,
		"limit": true, "data_source_id": dsi,
		"name":      qName,
		"raw_query": sql, "token": qToken}

	ireport := map[string]interface{}{"query": query}
	DoPVerb("PATCH", "reports/"+rToken+"/queries/"+qToken, ireport)

	thing := map[string]interface{}{"selected": true, "value": 100}
	rr := map[string]interface{}{"limit": thing}
	iqueries := []map[string]interface{}{query}

	report := map[string]interface{}{"name": "Api Test", "description": "",
		"report_run": rr,
		"queries[]":  iqueries,
		"trk_source": "editor"}
	ireport = map[string]interface{}{"report": report}
	DoPVerb("POST", "reports/"+rToken+"/runs", ireport)
	for {
		time.Sleep(750 * time.Millisecond)
		r := DoVerb("reports/" + rToken + "/queries/" + qToken + "/runs")
		latest := handleLinks(r, "query_runs", false)
		if latest != cra {
			break
		}
		fmt.Println(latest)
	}
	list := ReadList("query_runs")
	time.Sleep(750 * time.Millisecond)
	r = DoVerbFullPath(list[0] + "/content.json")
	fmt.Println(r)
}
func SqlAction(c *cli.Context) {
	rToken := ReadLast("report")
	qToken := ReadLast("query")

	queries := DoVerb("reports/" + rToken + "/queries")
	items := handleThing(queries, "queries", false)
	for _, item := range items {
		token, _ := item.GetString("token")
		sql, _ := item.GetString("raw_query")
		if token == qToken {
			SaveSQL(sql, token)
		}
	}
	path := UserHomeDir() + "/.mql_" + qToken + ".sql"
	cmd := exec.Command("vim", path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
}
func handleLinks(thing, meta string, print bool) string {
	v, _ := jason.NewObjectFromBytes([]byte(thing))
	if v == nil {
		return ""
	}
	e, _ := v.GetObject("_embedded")
	s, _ := e.GetObjectArray(meta)
	//token name
	list := []string{}
	cra := ""
	for i, item := range s {
		l, _ := item.GetObject("_links")
		r, _ := l.GetObject("result")
		href, _ := r.GetString("href")
		cra, _ = item.GetString("created_at")
		//coa, _ := item.GetString("completed_at")
		list = append(list, href)
		if print {
			tokens := strings.Split(cra, "T")
			fmt.Printf("%d. %s %s\n", i+1, tokens[0], strings.Split(tokens[1], ".")[0])
		}
		break
	}
	SaveList(meta, list)
	return cra
}
