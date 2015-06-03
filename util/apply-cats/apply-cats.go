package main

import (
	"flag"
	"strings"

	"github.com/golang/glog"
	"github.com/mattbaird/elastigo/lib"
	"gopkg.in/mgo.v2/bson"
)

type CatMap struct {
	Alias      string
	Categories []string
}

var catmap = []CatMap{
	CatMap{
		Alias: "Agricultural",
		Categories: []string{
			"Agricultural Products",
		},
	},
	CatMap{
		Alias: "Arts",
		Categories: []string{
			"Arts",
			"Charity",
			"Film",
			"Music",
			"Photography",
			"Society",
			"Theater",
		},
	},
	CatMap{
		Alias: "Automotive",
		Categories: []string{
			"Automobiles Components",
		},
	},
	CatMap{
		Alias: "Business",
		Categories: []string{
			"Business",
			"Company Site",
			"Human Resources",
			"Industrials",
			"Insurance",
			"International Agency",
			"Management",
			"Management Consulting",
			"Marketing",
			"Media",
			"Media Advertising",
			"Online Marketing",
			"Organisations",
			"Professional",
		},
	},
	CatMap{
		Alias: "Consumer",
		Categories: []string{
			"Beverages",
			"Consumables",
			"Consumer",
			"Extreme",
			"Fashion",
			"Food Products",
			"Gossip",
			"Home",
			"Hotels Restaurants Leisure",
			"Lifestyle",
			"Literature",
			"Offbeat",
			"Real Estate",
			"Recipes",
			"Recreation",
			"Religion",
			"Restaurants",
			"Retailing",
			"Satire",
			"Television",
			"Textiles Apparel",
			"Tobacco",
			"Travel Tourism",
			"Winter",
		},
	},
	CatMap{
		Alias: "Education",
		Categories: []string{
			"Education",
		},
	},
	CatMap{
		Alias: "Engineering and Science",
		Categories: []string{
			"Construction Engineering",
			"Chemicals",
			"Energy",
			"Environment",
			"Geographic",
			"Materials",
			"Metals Mining",
			"Oil Gas",
			"Paper Forest Products",
			"Transportation",
			"Utilities",
			"Science",
		},
	},
	CatMap{
		Alias: "Financial",
		Categories: []string{
			"Accounting",
			"Banking",
			"Financials",
		},
	},
	CatMap{
		Alias: "Government",
		Categories: []string{
			"Aerospace Defense",
			"Politics",
			"Government",
		},
	},
	CatMap{
		Alias: "Healthcare",
		Categories: []string{
			"Health",
			"Health Care",
			"Health Care Equipment Services",
			"Healthcare",
			"Medical",
			"Pharmaceuticals",
		},
	},
	CatMap{
		Alias: "Information Technology",
		Categories: []string{
			"Biotechnology",
			"Computers",
			"Electronic Equipment",
			"Imaging Equipment",
			"Software",
			"Telecommunication Services",
			"Information Technology",
			"Internet",
		},
	},
	CatMap{
		Alias: "Legal",
		Categories: []string{
			"Law",
		},
	},
	CatMap{
		Alias: "Logistics",
		Categories: []string{
			"Air Freight Logistics",
			"Containers Packaging",
		},
	},
	CatMap{
		Alias: "News",
		Categories: []string{
			"Global",
			"Local",
			"Miscellaneous",
			"National",
			"News",
			"Regional",
			"Standard",
			"Wires",
		},
	},
	CatMap{
		Alias: "Sports",
		Categories: []string{
			"American Football",
			"Athletics",
			"Badminton",
			"Baseball",
			"Basketball",
			"Boxing",
			"Casinos Gaming",
			"Cricket",
			"Cycling",
			"Field Hockey",
			"Fishing",
			"Games",
			"Golf",
			"Gymnastics",
			"Handball",
			"Horse Racing",
			"Ice Hockey",
			"Martial Arts",
			"Motor Racing",
			"Olympics",
			"Rowing",
			"Rugby",
			"Snooker",
			"Soccer",
			"Sports",
			"Squash",
			"Swimming",
			"Table Tennis",
			"Tennis",
			"Volleyball",
			"Wrestling",
			"Yachting",
		},
	},
}

func main() {
	flag.Parse()
	conn := elastigo.NewConn()
	conn.SetHosts([]string{"192.168.20.17", "192.168.20.18", "192.168.20.19"})
	settings := struct {
		Mappings bson.M `json:"mappings"`
	}{
		Mappings: bson.M{
			"alias": bson.M{
				"properties": bson.M{
					"Alias":      bson.M{"type": "string", "index": "not_analyzed"},
					"Categories": bson.M{"type": "string", "index": "not_analyzed"},
				},
			},
		},
	}

	if _, err := conn.CreateIndexWithSettings("categories", settings); err != nil {
		glog.Warningf("conn.CreateIndexWithSettings(): %s", err)
	}

	for _, cat := range catmap {
		for i := range cat.Categories {
			cat.Categories[i] = strings.Replace(cat.Categories[i], " ", "", -1)
		}
		_, err := conn.Index("categories", "alias", bson.NewObjectId().Hex(), nil, cat)
		if err != nil {
			glog.Fatalf("bulk.Index(): %s", err)
		}
	}
}
