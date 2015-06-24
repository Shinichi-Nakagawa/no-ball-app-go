package main

import (
	"database/sql"
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"html/template"
	"log"
	"math"
)

type Teams struct {
	Id     int64  `db:"yearID"`
	LgId   string `db:"lgID"`
	TeamID string `from:"teamID"`
	W      int64  `form:"W"`
	L      int64  `form:"L"`
	R      int64  `form:"R"`
	RA     int64  `form:"RA"`
	G      int64  `form:G`
	DivID  string `form:"divID"`
	Rank   string `form:"Rank"`
	Name   string `form:"name"`
}

type SearchForm struct {
	Year   string `form:"year" binding:"required"`
	League string `form:"league" binding:"required"`
}

type PythagorasForm struct {
	Years  []string
	League []string
}

type PythagorasResult struct {
	Form   PythagorasForm
	Year   string
	League string
	Teams  []Teams
}

func main() {

	// initialize the DbMap
	dbmap := initDb()
	defer dbmap.Db.Close()

	years := []string{"2014", "2013", "2012", "2011", "2010"}
	league := []string{"AL", "NL"}
	form := PythagorasForm{
		years,
		league,
	}
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Directory:  "templates",
		Layout:     "layout",
		Extensions: []string{".tmpl"},
		Charset:    "utf-8",
		Funcs: []template.FuncMap{
			{
				"calcPythagorasWin": func(r, ra, g int64) int64 {
					r2 := math.Pow(float64(r), 2)
					ra2 := math.Pow(float64(ra), 2)
					pythagorasPercent := r2 / (r2 + ra2)
					return int64(float64(g) * pythagorasPercent)
				},
			},
		},
	}))

	m.Get("/", func(r render.Render) {
		var teams []Teams
		result := PythagorasResult{
			form,
			"2014",
			"AL",
			teams,
		}
		r.HTML(200, "mlb/index", result)
	})

	m.Get("/pythagoras_search", binding.Bind(SearchForm{}), func(searchForm SearchForm, r render.Render) {
		var teams []Teams
		_, err := dbmap.Select(
			&teams,
			"select yearID, lgID, W, L, name, Rank, divID, teamID, R, RA, G from Teams where yearid = :year and lgid = :league ",
			map[string]interface{}{
				"year":   searchForm.Year,
				"league": searchForm.League,
			},
		)
		checkErr(err, "Select failed")
		result := PythagorasResult{
			form,
			searchForm.Year,
			searchForm.League,
			teams,
		}
		r.HTML(200, "mlb/pythagoras", result)
	})
	m.Run()
}

func initDb() *gorp.DbMap {
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish

	db, err := sql.Open("mysql", "app:adam_dunn@tcp(192.168.33.10:3306)/sean_lahman")
	checkErr(err, "sql.Open failed")

	// construct a gorp DbMap
	// dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}

	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
		panic(msg)
	}
}
