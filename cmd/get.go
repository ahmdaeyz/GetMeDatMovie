// Copyright © 2018 ahmdaeyz <ahmedalarabe5@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/spf13/cobra"
	bitly2 "github.com/zpnk/go-bitly"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type jsonurl struct {
	URL string `json:"url"`
}
type MovieQuality struct {
	quality string
	apicall string
}
type Query struct {
	Movies []struct {
		Title string `json:"t"`
		URL   string `json:"u"`
	} `json:"results"`
}

var movieQualities []MovieQuality

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Direct links to different qualities to the movie",
	Run: func(cmd *cobra.Command, args []string) {
		//quality,_:=cmd.Flags().GetString("quality")
		//TODO make getMovieLinks a map to get qualities easily
		b := bitly2.New(BitlyAccessToken)
		query := QuerySite(strings.Join(args, " "))
		if len(query.Movies) != 0 {
			movieLinks := GetMovieLinks("https://egy.best/" + query.Movies[0].URL + "/")
			for i, link := range movieLinks {
				fmt.Println(movieQualities[i].quality, ":")
				shrt, _ := b.Links.Shorten(link)
				fmt.Println(shrt.URL)
			}
			fmt.Println("Have fun!!")
		} else {
			fmt.Println("couldn't find", strings.Join(args, " "))
		}
	},
}

func init() {
	//Not yet Implemented
	getCmd.Flags().StringP("quality", "q", "available", "Specify movie quality")
	rootCmd.AddCommand(getCmd)
}
func GetJson(apicall string, movieName string) string {
	urlobj := new(jsonurl)
	spaceClient := http.Client{
		Timeout: time.Second * 120,
	}
	req, err := http.NewRequest(http.MethodGet, "https://egy.best/api?call="+apicall, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.75 Safari/537.36")
	req.Header.Set("Referer", "https://egy.best/movie/"+movieName+"/")
	req.Header.Set("Cookie", SiteCookie)
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	json.Unmarshal(body, urlobj)
	return urlobj.URL
}
func GetMovieLinks(url string) []string {
	var apiCalls []string
	var qualities []string
	var downloadLinks []string
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.75 Safari/537.36"),
	)
	//getting apicalls to qualities
	c.OnHTML("table.dls_table.btns.full.mgb tbody tr td.tar a.btn.g.dl.show_dl.api", func(element *colly.HTMLElement) {
		apiCalls = append(apiCalls, element.Attr("data-call"))
	})
	//getting qualities' names
	c.OnHTML("table.dls_table.btns.full.mgb tbody tr td", func(element *colly.HTMLElement) {
		if strings.Contains(element.Text, "p") {
			qualities = append(qualities, strings.Replace(strings.Replace(element.Text, "تحميل من EgyBest", "", -1), "  ", "", -1))
		}
	})
	c.Visit(url)
	for i := 0; i < len(apiCalls); i++ {
		movieQualities = append(movieQualities, MovieQuality{qualities[i], apiCalls[i]})
	}
	for _, movieQuality := range movieQualities {
		downloadLinks = append(downloadLinks, GetJson(movieQuality.apicall, func(string) (movieName string) {
			url = strings.Replace(url, "https://egy.best/movie/", "", -1)
			movieName = strings.Replace(url, "/", "", -1)
			return movieName
		}(url)))
	}
	return downloadLinks
}

func QuerySite(searchTerm string) *Query {
	client := http.Client{
		Timeout: 160 * time.Second,
	}
	req, err := http.NewRequest(http.MethodGet, "https://egy.best/autoComplete.php?q="+searchTerm, nil)
	if err != nil {
		log.Fatal(err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		if err != nil {
			log.Fatal(err)
		}
	}
	body, _ := ioutil.ReadAll(res.Body)
	str := string(body)
	str = strings.Replace(str, str[:strings.Index(str, "[")], "{\"results\":", -1)
	query := new(Query)
	json.Unmarshal([]byte(str), query)
	return query
}

func IsArabicMovie(searchTerm string) bool {
	var isArabic bool
	letters := []string{"ي", "و", "ه", "ن", "م", "ل", "ك", "ق", "ف", "غ", "ع", "ظ", "ط", "ض", "ص", "ش", "س", "ز", "ر", "ذ", "د", "خ", "ح", "ج", "ث", "ت", "ب", "ا"}
	for _, letter := range letters {
		if strings.Contains(searchTerm, letter) {
			isArabic = true
			break
		}
	}
	return isArabic
}
