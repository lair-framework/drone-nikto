package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/lair-framework/api-server/client"
	"github.com/lair-framework/go-lair"
	"github.com/lair-framework/go-nikto"
)

const (
	version = "2.0.1"
	tool    = "nikto"
	usage   = `
Usage:
  drone-nikto <id> <filename>
  export LAIR_ID=<id>; drone-nikto <filename>
Options:
  -v              show version and exit
  -h              show usage and exit
  -k              allow insecure SSL connections
  -force-ports    disable data protection in the API server for excessive ports
  -tags           a comma separated list of tags to add to every host that is imported
`
)

func buildProject(nikto *nikto.NiktoData, exproject *lair.Project, tags []string) (map[string]bool, error) {
	bNotFound := make(map[string]bool)
	exproject.Tool = tool
	var command string

	for _, scan := range nikto.NiktoScan {
		for _, item := range scan.ScanDetails {
			command = scan.Options

			// Confirm if Host and Port existing in current project
			found := false
			hname := false
			for x, h := range exproject.Hosts {
				if item.TargetIP == h.IPv4 {
					for y, p := range h.Services {
						if item.TargetPort == p.Port {
							found = true

							// Append host tags
							exproject.Hosts[x].Tags = append(h.Tags, tags...)

							// Update Modified By
							h.LastModifiedBy = tool
							p.LastModifiedBy = tool

							// Append Note information to Service
							note := &lair.Note{
								Title:          fmt.Sprintf("Nikto v%v (%v:%v)", scan.Version, item.HostHeader, item.TargetPort),
								LastModifiedBy: tool,
							}

							scheme, err := url.Parse(item.TargetHostname)
							if err != nil {
								return bNotFound, err
							}
							if scheme.Scheme == "https" {
								note.Content = fmt.Sprintf("SSL Information:\nSubject: %v\nCiphers: %v\nIssuer: %v\n\n", item.SSL.Info, item.SSL.Ciphers, item.SSL.Issuers)
							}

							for _, item := range item.Items {
								if item.OSVDBID > 0 {
									note.Content += fmt.Sprintf("%v URI: %v OSVDBID: %v\n", item.Description, item.URI, item.OSVDBID)
								} else {
									note.Content += fmt.Sprintf("%v URI: %v\n", item.Description, item.URI)
								}
							}
							note.Content += fmt.Sprintf("\nStart: %s End: %s", scan.ScanStart, scan.ScanEnd)
							exproject.Hosts[x].Services[y].Notes = append(p.Notes, *note)
						}
					}
					// Check for existing hostname and append
					for _, n := range h.Hostnames {
						if n == item.TargetHostname {
							hname = true
						}
					}
					if !hname && (item.TargetHostname != item.TargetIP) {
						exproject.Hosts[x].Hostnames = append(h.Hostnames, item.TargetHostname)
					}
				}
			}
			if !found {
				bNotFound[fmt.Sprintf("%s (%s:%v)", item.TargetHostname, item.TargetIP, item.TargetPort)] = true
				continue
			}
		}
	}
	// Update project with Nikto command
	var com []lair.Command
	com = append(com, lair.Command{
		Tool:    tool,
		Command: command,
	})
	exproject.Commands = com

	return bNotFound, nil
}

func main() {
	showVersion := flag.Bool("v", false, "")
	insecureSSL := flag.Bool("k", false, "")
	forcePorts := flag.Bool("force-ports", false, "")
	tags := flag.String("tags", "", "")
	flag.Usage = func() {
		fmt.Println(usage)
	}
	flag.Parse()
	if *showVersion {
		log.Println(version)
		os.Exit(0)
	}
	lairURL := os.Getenv("LAIR_API_SERVER")
	if lairURL == "" {
		log.Fatal("Fatal: Missing LAIR_API_SERVER environment variable")
	}
	lairPID := os.Getenv("LAIR_ID")
	var filename string
	switch len(flag.Args()) {
	case 2:
		lairPID = flag.Arg(0)
		filename = flag.Arg(1)
	case 1:
		filename = flag.Arg(0)
	default:
		log.Fatal("Fatal: Missing required argument")
	}

	u, err := url.Parse(lairURL)
	if err != nil {
		log.Fatalf("Fatal: Error parsing LAIR_API_SERVER URL. Error %s", err.Error())
	}
	if u.User == nil {
		log.Fatal("Fatal: Missing username and/or password")
	}
	user := u.User.Username()
	pass, _ := u.User.Password()
	if user == "" || pass == "" {
		log.Fatal("Fatal: Missing username and/or password")
	}
	c, err := client.New(&client.COptions{
		User:               user,
		Password:           pass,
		Host:               u.Host,
		Scheme:             u.Scheme,
		InsecureSkipVerify: *insecureSSL,
	})
	if err != nil {
		log.Fatalf("Fatal: Error setting up client. Error %s", err.Error())
	}

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Fatal: Could not open file. Error %s", err.Error())
	}

	niktoData, err := nikto.Parse(buf)
	if err != nil {
		log.Fatalf("Fatal: Error parsing nikto data. Error %s", err.Error())
	}

	hostTags := []string{}
	if *tags != "" {
		hostTags = strings.Split(*tags, ",")
	}

	project, err := c.ExportProject(lairPID)
	if err != nil {
		log.Fatalf("Fatal: Unable to export project. Error %s", err.Error())
	}
	bNotFound, err := buildProject(niktoData, &project, hostTags)
	if err != nil {
		log.Fatal(err.Error())
	}

	res, err := c.ImportProject(&client.DOptions{ForcePorts: *forcePorts}, &project)
	if err != nil {
		log.Fatalf("Fatal: Unable to import project. Error %s", err)
	}

	defer res.Body.Close()
	droneRes := &client.Response{}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Fatal: Error %s", err.Error())
	}
	if err := json.Unmarshal(body, droneRes); err != nil {
		log.Fatalf("Fatal: Could not unmarshal JSON. Error %s", err.Error())
	}
	if droneRes.Status == "Error" {
		log.Fatalf("Fatal: Import failed. Error %s", droneRes.Message)
	}

	if len(bNotFound) > 0 {
		log.Printf("Info: The following host ports contained Nikto results but did not exist in project %s", lairPID)
	}
	for k := range bNotFound {
		fmt.Println(k)
	}

	log.Println("Success: Operation completed successfully")
}
