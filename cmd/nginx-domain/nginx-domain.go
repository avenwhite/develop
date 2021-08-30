package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"mux"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"
)
type DomainRequest  struct {
	DomainName      		string 			`json:"name"`
	Host                	string 			`json:"host"`
	Port	                string        	`json:"port"`
}
type DomainResponse struct {
	Name string `json:"name"`
}
var domain []DomainRequest

func sliceToDomainRequest(elements []string) {
	elementMap := make(map[string]string)
	domain = nil
	for i:=0; i<len(elements)-1; i+=2  {
		elementMap["DomainName"] = elements[i]
		var re = regexp.MustCompile(`http:\/\/(.*):`)
		matchHost := re.FindStringSubmatch(elements[i+1])
		elementMap["Host"] = matchHost[1]
		re = regexp.MustCompile(`http:\/\/.*:(.*$)`)
		matchPort := re.FindStringSubmatch(elements[i+1])
		elementMap["Port"] = matchPort[1]
		domain = append(domain, DomainRequest{DomainName: elementMap["DomainName"], Host: elementMap["Host"], Port: elementMap["Port"]})
		//fmt.Println(domain)
	}
}
func readConfFile(fileName string) []string{
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	var fileConf string
	fileConf=string(bytes[:])
	re:= regexp.MustCompile(`(?sm)(server .*?}+.})`)
	matchServers := re.FindAllString(fileConf,-1)
	return matchServers
}
func saveConfFile(fileName string,input []string) {
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	for _,s:=range input{
		f.WriteString(s+"\n")
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func deleteConfFile(name string) {
	tempS:= readConfFile("files/nginx.conf")
	for i,item:= range tempS {
		if strings.Contains(item, name){
			tempS=append(tempS[:i], tempS[i+1:]...)
			}
	}
	saveConfFile("files/nginx.conf",tempS)
}

func parseConfFile() {
	var fileContent []string
	var tempS [][]string
	for _,item:= range readConfFile("files/nginx.conf") {
		re := regexp.MustCompile(`(proxy_pass|server_name)(.*);`)
		tempS = re.FindAllStringSubmatch(item,-1)
		fileContent = append(fileContent,tempS[0][2],tempS[1][2])
	}
	sliceToDomainRequest(fileContent)
}

func deleteDomain(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	parseConfFile()
	var tdomain DomainResponse
	var flag bool
	_ = json.NewDecoder(r.Body).Decode(&tdomain)
	for _,item:= range domain {
		if strings.Contains(item.DomainName, tdomain.Name) {
			flag=true
		} else {
			flag=false
		}
	}
	if flag {
		deleteConfFile(tdomain.Name)
	} else {
		fmt.Fprint(w, "Error, domain id not exist")
	}
}

func getDomains(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	parseConfFile()
	json.NewEncoder(w).Encode(domain)
//	log.Println("getDomains: ",domain)
}

func saveDomain(tdomain DomainRequest,filename string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	t, err := template.New("domains").Parse("\nserver {\n    listen 80;\n    server_name {{ .DomainName}};\n    location / {\n        proxy_pass http://{{ .Host}}:{{ .Port}};\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n        proxy_set_header X-Forwarded-Proto $scheme;\n    }\n}")
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, tdomain)

	if _, err = f.WriteString(buf.String()); err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}

}

func createDomain (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		parseConfFile()
		var tdomain DomainRequest
		var flag bool
		_ = json.NewDecoder(r.Body).Decode(&tdomain)
		for _,item:= range domain {
			if strings.Contains(item.DomainName, tdomain.DomainName) {
				flag=true
			} else {
				flag=false
			}
		}
		if flag {
			fmt.Fprint(w, "Error, domain id already exist")
		} else {
			domain = append(domain, tdomain)
			json.NewEncoder(w).Encode(tdomain)
			saveDomain(tdomain,"files/nginx.conf")
		}
}

func main() {
	cfgPath, err := ParseFlags()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/domains", getDomains).Methods("GET")
	r.HandleFunc("/domain", createDomain).Methods("POST")
	r.HandleFunc("/domain", deleteDomain).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8000", r))
}
