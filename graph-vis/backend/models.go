package main

type TaxonBase struct {
	Rank string `json:"rank"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Taxon struct {
	TaxonBase
	Id string `json:"_id"`
}

type TaxonResponse struct {
	TaxonBase
	Id string `json:"id"`
}
