package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
)

type JourResponse struct {
	ResponseData struct {
		Result []JourneyPatternPointOnLine `json:"Result"`
	} `json:"ResponseData"`
}

type JourneyPatternPointOnLine struct {
	LineNumber                string `json:"LineNumber"`
	JourneyPatternPointNumber string `json:"JourneyPatternPointNumber"`
}

type StopResponse struct {
	ResponseData struct {
		Result []StopPoint `json:"Result"`
	} `json:"ResponseData"`
}

type StopPoint struct {
	StopPointNumber string `json:"StopPointNumber"`
	StopPointName   string `json:"StopPointName"`
}

type LineResponse struct {
	ResponseData struct {
		Result []Line `json:"Result"`
	} `json:"ResponseData"`
}

type Line struct {
	LineNumber      string `json:"LineNumber"`
	LineDesignation string `json:"LineDesignation"`
}

type LineStop struct {
	LineNumber      string      `json:"lineNumber"`
	LineDesignation string      `json:"lineDesignation"`
	Stops           []StopPoint `json:"stops"`
}

type ByStops []LineStop

func (a ByStops) Len() int           { return len(a) }
func (a ByStops) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByStops) Less(i, j int) bool { return len(a[i].Stops) > len(a[j].Stops) }

func fetch(model, key string, dest interface{}) {
	url := fmt.Sprintf("https://api.sl.se/api2/LineData.json?model=%v&key=%v&DefaultTransportModeCode=BUS", model, key)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Accept-Encoding", "gzip,deflate")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("expected HTTP OK, got %v\n", resp.StatusCode)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	if err := json.Unmarshal(b, dest); err != nil {
		log.Fatalln(err)
	}
}

func jours(key string, stopPoints, lines map[string]string) []LineStop {
	jr := JourResponse{}
	fetch("jour", key, &jr)
	lineStops := map[string]LineStop{}
	for _, l := range jr.ResponseData.Result {
		if _, found := lineStops[l.LineNumber]; !found {
			lineStops[l.LineNumber] = LineStop{
				LineNumber:      l.LineNumber,
				LineDesignation: lines[l.LineNumber],
				Stops:           []StopPoint{},
			}
		}
		lineStop := lineStops[l.LineNumber]
		lineStop.Stops = append(lineStop.Stops, StopPoint{
			StopPointNumber: l.JourneyPatternPointNumber,
			StopPointName:   stopPoints[l.JourneyPatternPointNumber],
		})
		lineStops[l.LineNumber] = lineStop
	}
	result := []LineStop{}
	for _, v := range lineStops {
		result = append(result, v)
	}
	sort.Sort(ByStops(result))
	return result[:10]
}

func stopPoints(key string) map[string]string {
	ld := StopResponse{}
	fetch("stop", key, &ld)
	sp := map[string]string{}
	for _, s := range ld.ResponseData.Result {
		sp[s.StopPointNumber] = s.StopPointName
	}
	return sp
}

func lines(key string) map[string]string {
	lr := LineResponse{}
	fetch("line", key, &lr)
	sp := map[string]string{}
	for _, l := range lr.ResponseData.Result {
		sp[l.LineNumber] = l.LineDesignation
	}
	return sp
}

func main() {
	key := os.Getenv("API_KEY")
	if key == "" {
		log.Fatalln("enovironment variable API_KEY is empty, set to Trafiklab API Key")
	}
	log.Println("fetching data")
	sp := stopPoints(key)
	lines := lines(key)
	jours := jours(key, sp, lines)

	log.Println("starting web server")
	http.Handle("/", http.FileServer(http.Dir("/www")))
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		j, err := json.Marshal(jours)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(j)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
