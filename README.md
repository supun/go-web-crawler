# go-web-crawler
simple go web server to scrape web pages using URLs

##  steps
1. Clone the project
2. Get below dependencies 
  ```bash
  go get -u github.com/labstack/echo/v4
  go get -u github.com/lestoni/html-version
  ```
3. run ```bash go run server.go``

## Assumptions
1. If any page contains password type input and subnit type input tag, it is cosidered the page has a login form
2. Subdomain links are considered as internal links.

## Improvements
1. Adding a configurable server time-out to handle unreachable links.
2. Make UI more user friendly 
3. Handle multiple URL checking requests parallely.
