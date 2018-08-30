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
		"github.com/spf13/cobra"
	"strings"
	"net/http"
	"time"
	"log"
	"io/ioutil"
	"encoding/json"
	"github.com/gocolly/colly"
	"fmt"
	"github.com/StalkR/imdb"
	"strconv"
	bitly2 "github.com/zpnk/go-bitly"
)
type jsonurl struct {
	URL     string `json:"url"`
}
type MovieQuality struct {
	quality string
	apicall string
}

var movieQualities []MovieQuality
// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Direct links to different qualities to the movie",
	Run: func(cmd *cobra.Command, args []string) {
		//quality,_:=cmd.Flags().GetString("quality")
		//TODO make getMovieLinks a map to get qualities easily
		//TODO handle commas and dots
		imdbClient:=&http.Client{
			Timeout:time.Second*60,
		}
		b:=bitly2.New("3b9f54331154f15b052980bbcd812009746b4533")
		results,err:=imdb.SearchTitle(imdbClient,strings.Join(args," "))
		if err!=nil{
			log.Fatal(err)
		}
		if len(results)!=0 {
			movieName := strings.Replace(strings.Replace(strings.ToLower(results[0].Name), " ", "-", -1),":","",-1) + "-" + strconv.Itoa(results[0].Year)
			for i, link := range GetMovieLinks("https://egy.best/movie/" + movieName + "/") {
				fmt.Println(movieQualities[i].quality, ":")
				shrt,_:=b.Links.Shorten(link)
				fmt.Println(shrt.URL)
			}
			fmt.Println("Have fun!!")
		}else{
			fmt.Println("Couldn't find this title!")
		}
	},
}

func init() {
	getCmd.Flags().StringP("quality","q","available","Specify movie quality")
	rootCmd.AddCommand(getCmd)
}
func GetJson(apicall string,movieName string) string{
	urlobj:=new(jsonurl)
	spaceClient := http.Client{
		Timeout: time.Second * 30,
	}
	req, err := http.NewRequest(http.MethodGet, "https://egy.best/api?call="+apicall, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.75 Safari/537.36")
	req.Header.Set("Referer","https://egy.best/movie/"+movieName+"/")
	req.Header.Set("Cookie","__cfduid=d0e60ae5885c4cab4d4ea0679036fda321535226127; PSSID=80u45JciUr6YD7mBNDc3k9UE0JJeBgiv5UbLJpvzfFrVaBMOWEinrFfrTnBtgTH5bO5YriiG4x5F6--COFPT53; _ga=GA1.2.1978196420.1535226077; _gid=GA1.2.1079465582.1535226077; __test; JS_TIMEZONE_OFFSET=-10800; 494668b4c0ef4d25bda4e75c27de2817=1b9c99d8-c3ec-4cf1-bb7d-01e9a962fcc3:2:2; adcashufpv3=131842452215161033441756794336; ppu_main_edd345426cf699c97a2fbda6dee85a60=1; EGUserDef=1122600%7C203b1b642b33933b83db3652111d7b93%7C812107a88d681c01b556f2404b64f2bb; noprpkedvhozafiwrcnt=1; noprpkedvhozafiwrexp=Tue, 28 Aug 2019 22:52:27 GMT; _gat_gtag_UA_9923694_2=1; ppu_sub_edd345426cf699c97a2fbda6dee85a60=2; ppu_delay_edd345426cf699c97a2fbda6dee85a60=1")
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	json.Unmarshal(body,urlobj)
	return urlobj.URL
}
func GetMovieLinks(url string) []string{
	var apiCalls []string
	var qualities []string
	var downloadLinks []string
	c:=colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.75 Safari/537.36"),
	)
	//c.SetProxy("//185.206.125.38:80")
	//getting apicalls to qualities
	c.OnHTML("table.dls_table.btns.full.mgb tbody tr td.tar a.btn.g.dl.show_dl.api", func(element *colly.HTMLElement) {
		apiCalls=append(apiCalls,element.Attr("data-call"))
	})
	//getting qualities' names
	c.OnHTML("table.dls_table.btns.full.mgb tbody tr td", func(element *colly.HTMLElement) {
		if strings.Contains(element.Text,"p") {
			qualities = append(qualities, strings.Replace(strings.Replace(element.Text, "تحميل من EgyBest", "", -1), "  ", "", -1))
		}
	})
	c.Visit(url)
	for i:=0;i< len(apiCalls);i++{
		movieQualities=append(movieQualities, MovieQuality{qualities[i],apiCalls[i]})
	}
	for _,movieQuality:=range movieQualities {
		downloadLinks= append(downloadLinks,GetJson(movieQuality.apicall, func(string) (movieName string ) {
			url=strings.Replace(url,"https://egy.best/movie/","",-1)
			movieName=strings.Replace(url,"/","",-1)
			return movieName
		}(url)))
	}
	return downloadLinks
}