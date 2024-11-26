package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type newsprdct struct {
	Id        int           `json:"id"`
	Link      string        `json:"link"`
	Date      string        `json:"date"`
	Source    string        `json:"source"`
	News      string        `json:"news"`
	News_text string        `json:"news_text"`
	Products  []productprob `json:"products"`
}

type product struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
}

type productprob struct {
	Id          int     `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Prob        float32 `json:"prob"`
	Just        string  `json:"just"`
}

func news_prod_f(news_text string, renew bool) []productprob {
	products, err := os.ReadFile("./data/products.json")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	var prds []product
	json.NewDecoder(bytes.NewBuffer(products)).Decode(&prds)

	prprbsl := []productprob{}
	for i, p := range prds {

		tprd_descr := p.Title + " " + p.Description
		prprb := productprob{}
		prprb.Id = i
		prprb.Title = p.Title
		prprb.Description = p.Description
		prprb.Prob = 1

		if renew {
			prprb.Prob, prprb.Just = Get_prob_open_ai(news_text, tprd_descr)
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
		} else {
			prprb.Prob, prprb.Just = get_prob(news_text, tprd_descr)
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
		}
		prprbsl = append(prprbsl, prprb)
		fmt.Println("Sleep 2, ", news_text[:10], p.Title)
		time.Sleep(time.Second * 2)
	}
	return prprbsl
}

func item_Title_in_news_prod(title string, news_prod []newsprdct) (isinnews bool) {
	isinnews = false
	for _, nw := range news_prod {
		if title == nw.News {
			isinnews = true
			break
		}
	}
	return
}

// func Pars(news_prod *[]newsprdct) {
func Pars(news_prod []newsprdct, renew bool) []newsprdct {
	maxnews := 10
	source := "https://www.mskagency.ru"
	file, _ := os.Open("./data/building.rss")
	defer file.Close()
	fp := gofeed.NewParser()
	feed, _ := fp.Parse(file)
	log.Println("Всего на сайте новостей: ", len(feed.Items))
	// news_prod := []newsprdct{}
	log.Printf("Цикл обработки rss, новостей в json перед циклом: %d", len(news_prod))
	// for _, item := range feed.Items[:10] {
	cntnews := 0
	for i := len(feed.Items); i > 0; i-- {
		if i <= maxnews {
			item := feed.Items[i-1]
			if !item_Title_in_news_prod(item.Title, news_prod) {

				re := regexp.MustCompile(`\d{2}\.\d{2}\.\d{4} \d{2}:\d{2}`)
				dateTime := re.FindString(item.Description)

				news_text_1 := strings.ReplaceAll(item.Description, `. Агентство "Москва". `, "")
				news_text_2 := strings.ReplaceAll(news_text_1, dateTime, "")
				retag := regexp.MustCompile(`<.*?>`)
				news_text := retag.ReplaceAllString(news_text_2, "")

				reid := regexp.MustCompile(`\d{7}`)
				idnews, err := strconv.Atoi(reid.FindString(item.Link))
				if err != nil {
					log.Println("Ошибка в получении idnews", err)
					os.Exit(1)
				}

				newprd := newsprdct{
					Id:        idnews,
					Link:      item.Link,
					Date:      dateTime,
					Source:    source,
					News:      item.Title,
					News_text: news_text,
					Products:  news_prod_f(news_text, renew),
				}

				// *news_prod = append(*news_prod, newprd)
				news_prod = append([]newsprdct{newprd}, news_prod...)
				log.Println("+: ", string(item.Title))
				cntnews++
			}
		}
	}
	jstsv, err := json.Marshal((news_prod)[:maxnews])
	if err == nil {
		os.WriteFile("./data/news_prod.json", jstsv, 0644)
		log.Printf("Добавлено %d новостей.", cntnews)
	}

	return news_prod
}

func Check_news() {
	news_prod_file, err := os.ReadFile("./data/news_prod.json")
	news_prod := []newsprdct{}
	if err == nil {
		json.NewDecoder(bytes.NewBuffer(news_prod_file)).Decode(&news_prod)
	} else {
		log.Println("Нет файла news_prod.json, создание заново.")
	}
	// Pars(&news_prod)
	renew := false
	news_prod = Pars(news_prod, renew)
}

func Check_news_renew() {
	renew := false
	news_prod := []newsprdct{}
	Pars(news_prod, renew)
}
