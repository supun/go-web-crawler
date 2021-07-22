package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/lestoni/html-version"
	"golang.org/x/net/html"
)

type (
	FormData struct {
		URL   string    `json:"url"`
	}

	Heading struct{
		Type string `json:"type"`
		Count int `json:"count"`
	}

	HtmlContentResponse struct{
		StatusMessage string `json:"status_message"`
		Status int `json:"status"`
		Version string `json:"version"`
		Title string `json:"title"`
		Headings []Heading `json:"headings"`
		InternalLinks int `json:"internal_links"`
		ExternalLinks int `json:"external_links"`
        InaccessibleLinks int `json:"inaccessible_links"`
		HasLoginForm bool `json:"has_login_form"`
		Links [] string `json:"links"`
	}
)
func main() { 
	e := echo.New()
	
	// Routes
	e.File("/","public/index.html")
	e.POST("/api/check-url", checkUrl)
	e.Logger.Fatal(e.Start(":1323"))
}

func checkUrl(c echo.Context) error{
	formdata := new(FormData)
	if err := c.Bind(formdata); err != nil {
		return err
	}

	if formdata.URL == ""{
		return c.JSON(http.StatusOK,HtmlContentResponse{
			Status: http.StatusBadRequest,
			StatusMessage :http.StatusText(http.StatusBadRequest),
		})
	}
	
	page, err := parse(formdata.URL)
	if err != nil {
		return c.JSON(http.StatusOK,HtmlContentResponse{
			Status: http.StatusNotFound,
			StatusMessage :http.StatusText(http.StatusNotFound),
		})
	}
	title := getPageTitle(page)
	var links [] string
	var inputs [] * html.Node
	u, err := url.Parse(formdata.URL)
	if err != nil {
			log.Fatal(err)
	}
	parts := strings.Split(u.Hostname(), ".")
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]

    links = getPageLinks(links,page)
	
	internallinks,externallinks := processLinks(links,domain)
	invalidlInks := getInvalidLinks(links,formdata.URL)
    inputs = getInputs(inputs,page)
	loginformfound := checkLoginForm(inputs)
	version, err := HTMLVersion.DetectFromURL(formdata.URL)
	
	if err != nil{
		return c.JSON(http.StatusOK,HtmlContentResponse{
			Status: http.StatusNotFound,
		})
	}

	headings, err := getHeadings(formdata.URL)
	return c.JSON(http.StatusOK,HtmlContentResponse{
		Version: version,
		Status: http.StatusOK,
		Title: title,
		Headings: headings,
		Links: links,
		ExternalLinks: externallinks,
		InternalLinks: internallinks,
		InaccessibleLinks: invalidlInks,
		HasLoginForm: loginformfound,
	})
}

func parse(url string) (*html.Node, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Cannot get page")
	}
	b, err := html.Parse(r.Body)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse page")
	}
	return b, err
}

func getPageTitle(n *html.Node) string {
	var title string
	if n.Type == html.ElementNode && n.Data == "title" {
		return n.FirstChild.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		title = getPageTitle(c)
		if title != "" {
			break
		}
	}
	return title
}

func getHeadings(url string)([]Heading, error){
	response, err := http.Get(url)
	if err != nil {
		return nil,err
		}

tag := ""
	
tokenizer := html.NewTokenizer(response.Body)

var heading1,heading2,heading3,heading4,heading5,heading6 Heading 
for {
	tt := tokenizer.Next()
	token := tokenizer.Token()

	err := tokenizer.Err()
	if err == io.EOF {
		break
	}

	switch tt {
	case html.ErrorToken:
		log.Fatal(err)
	case html.StartTagToken:
		heading1.Type = "h1"
		heading2.Type = "h2"
		heading3.Type = "h3"
		heading4.Type = "h4"
		heading5.Type = "h5"
		heading6.Type = "h6"
		tag = token.Data
		switch tag {
		case "h1":
			heading1.Count++
		case "h2":
			heading2.Count++
		case "h3":
			heading3.Count++
		case "h4":
			heading4.Count++
		case "h5":
			heading5.Count++
		case "h6":
			heading6.Count++
		default:
		//	fmt.Println(data)
		}
		
	
	}
}

var headings []Heading
headings = append(headings,heading1,heading2,heading3,heading4,heading5,heading6)
return headings,nil

}


func getPageLinks(links []string, n *html.Node) []string {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" {
				if !sliceContains(links, a.Val) {
					links = append(links, a.Val)
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = getPageLinks(links, c)
	}
	return links
}


func getInputs(inputs []*html.Node, n *html.Node) []*html.Node {
	if n.Type == html.ElementNode && n.Data == "input" {
		inputs = append(inputs, n)
		
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		inputs = getInputs(inputs, c)
	}
	return inputs
}


func checkLoginForm(inputs []*html.Node) bool{
	var passwordfound, submitfound bool
	for _,v := range inputs{
		for _, a := range v.Attr {
			if a.Key == "type" && a.Val == "password" {
				passwordfound = true
			}

			if a.Key == "type" && a.Val == "submit" {
				submitfound = true
			}
		}
	
	}
	return passwordfound && submitfound
}

func processLinks(links [] string, domian string) (int,int){
   var internallinks,externallinks int
	for _,v := range links{
		if (strings.Contains(v,"http") || strings.Contains(v,"https") ) && !strings.Contains(v,domian){
			externallinks ++
		}else {
			internallinks ++
		} 
	}
return internallinks,externallinks
}

func getInvalidLinks(links []string, domain string) int{
	var invalidLinks int
	var link string
	for _,v := range links{
		link = v
		if !strings.Contains(v,"http") || !strings.Contains(v,"https") {
			link = domain+v
		}
		_, err := http.Get(link)
	if err != nil {
		invalidLinks++
	}
	}
	return invalidLinks
}

func sliceContains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}