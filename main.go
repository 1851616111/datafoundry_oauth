package main

import "net/http"

func main() {

	http.HandleFunc("/v1/github-redirect", githubHandler)
	http.ListenAndServe(":9443",nil)

}
