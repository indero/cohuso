package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Feed struct {
	XMLName    xml.Name `xml:"feed"`
	Text       string   `xml:",chardata"`
	Lang       string   `xml:"lang,attr"`
	Xmlns      string   `xml:"xmlns,attr"`
	OpenSearch string   `xml:"openSearch,attr"`
	Tel        string   `xml:"tel,attr"`
	ID         string   `xml:"id"`
	Title      struct {
		Text string `xml:",chardata"`
		Type string `xml:"type,attr"`
	} `xml:"title"`
	Generator struct {
		Text    string `xml:",chardata"`
		Version string `xml:"version,attr"`
		URI     string `xml:"uri,attr"`
	} `xml:"generator"`
	Updated string `xml:"updated"`
	Link    []struct {
		Text string `xml:",chardata"`
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
		Type string `xml:"type,attr"`
	} `xml:"link"`
	TotalResults string `xml:"totalResults"`
	StartIndex   string `xml:"startIndex"`
	ItemsPerPage string `xml:"itemsPerPage"`
	Query        struct {
		Text        string `xml:",chardata"`
		Role        string `xml:"role,attr"`
		SearchTerms string `xml:"searchTerms,attr"`
		StartPage   string `xml:"startPage,attr"`
	} `xml:"Query"`
	Image struct {
		Text   string `xml:",chardata"`
		Height string `xml:"height,attr"`
		Width  string `xml:"width,attr"`
		Type   string `xml:"type,attr"`
	} `xml:"Image"`
	Entry []struct {
		Text      string `xml:",chardata"`
		ID        string `xml:"id"`
		Updated   string `xml:"updated"`
		Published string `xml:"published"`
		Title     struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"title"`
		Content struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"content"`
		Nopromo string `xml:"nopromo"`
		Author  struct {
			Text string `xml:",chardata"`
			Name string `xml:"name"`
		} `xml:"author"`
		Link []struct {
			Text  string `xml:",chardata"`
			Href  string `xml:"href,attr"`
			Title string `xml:"title,attr"`
			Rel   string `xml:"rel,attr"`
			Type  string `xml:"type,attr"`
		} `xml:"link"`
	} `xml:"entry"`
}

type Snom struct {
	XMLName xml.Name `xml:"SnomIPPhoneText"`
	Title   string
	Prompt  string
	Text    string
}

func getXML(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("read body: %v", err)
	}

	return data, nil
}

func askTelSearch(number string) (*Feed, error) {
	if xmlBytes, err := getXML("https://tel.search.ch/api/?was=" + number + "&key=" + Api_Key); err != nil {
		log.Printf("Failed to get XML: %v", err)
		return &Feed{}, err
	} else {
		result := Feed{}
		xml.Unmarshal(xmlBytes, &result)
		return &result, nil
	}
}

func callHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/call.php" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	query := r.URL.Query()
	caller := query.Get("caller")
	if caller == "" {
		fmt.Println("Caller was empty")
		fmt.Fprintf(w, "Caller was empty")
		return
	}

	feed, _ := askTelSearch(caller)

	identified_caller := Snom{
		XMLName: xml.Name{Local: "Person"},
		Title:   "Anruf von " + caller,
		Prompt:  "Anruf von " + caller,
		Text:    "Keinen Eintrag gefunden",
	}

	// Found Entry in Directory
	if len(feed.Entry) == 1 {
		identified_caller.Text = feed.Entry[0].Content.Text
	}

	// Did not find a direct match, will start to remove digits
	i := 1
	for i = 1; len(feed.Entry) < 1 && i < 4; i++ {
		feed, _ = askTelSearch(caller[:len(caller)-i])
	}

	// We have at least one Entry in the Directory and
	// the digits remove counter is higher than one.
	if len(feed.Entry) > 0 && i > 1 {
		finalText := ""
		for i := 0; i < len(feed.Entry); i++ {
			finalText += feed.Entry[i].Title.Text + "</br>"
		}
		res := fmt.Sprintf("Anruf von %s (%d Resultate)", caller, len(feed.Entry))
		identified_caller.Title = res
		identified_caller.Prompt = res
		identified_caller.Text = finalText
	}

	x, err := xml.MarshalIndent(identified_caller, "", "	")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}

var (
	Api_Key         string
	HTTP_ListenAddr string = ":8080"
)

func main() {
	flag.StringVar(&Api_Key, "api-key", LookupEnvOrString("TEL_SEARCH_CH_API_KEY", Api_Key), "tel.search.ch api Key")
	flag.StringVar(&HTTP_ListenAddr, "http-listen-addr", LookupEnvOrString("HTTP_LISTEN_ADDR", HTTP_ListenAddr), "http service listen address")

	flag.Parse()
	log.Printf("app.config %v\n", getConfig(flag.CommandLine))

	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/call.php", callHandler)

	fmt.Printf("Starting server at %s\n", HTTP_ListenAddr)
	if err := http.ListenAndServe(HTTP_ListenAddr, nil); err != nil {
		log.Fatal(err)
	}
}

func LookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getConfig(fs *flag.FlagSet) []string {
	cfg := make([]string, 0, 10)
	fs.VisitAll(func(f *flag.Flag) {
		cfg = append(cfg, fmt.Sprintf("%s:%q", f.Name, f.Value.String()))
	})

	return cfg
}
