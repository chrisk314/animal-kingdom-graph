package main

type Taxon struct {
	Id   string `json:"_id"`
	Rank string `json:"rank"`
	Name string `json:"name"`
	Url  string `json:"url"`
}
