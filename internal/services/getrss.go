package services

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	"golang.org/x/exp/rand"
)

func Dlt_to_nxt_wrk_day(dt_tm time.Time) (dlt int) {
	file, err := os.Open("./data/calendar2024.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	hol := []string{}
	for scanner.Scan() {
		hol = append(hol, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	dlt = 0
	for {
		dt_tm_str := dt_tm.Format("2006.01.02")
		idx := slices.IndexFunc(hol, func(c string) bool { return c == dt_tm_str })
		if idx >= 0 {
			dt_tm = dt_tm.Add(time.Hour * 24)
			dlt++
		} else {
			break
		}
	}

	return
}

func Time_to_sleep_f(timeout int) (nexttime time.Time, time_to_sleep time.Duration) {
	beg_tm := 8 //Учесть часовой пояс сервера
	end_time := 17
	dt_tm := time.Now() //.Add(time.Hour * 24)
	dlt := Dlt_to_nxt_wrk_day(dt_tm)
	if dlt == 0 && dt_tm.Hour() >= end_time {
		dlt = 1
	}
	dt_tm_00 := time.Date(dt_tm.Year(), dt_tm.Month(), dt_tm.Day(), 0, 0, 0, 0, time.Local)

	switch {
	case dlt == 0:
		nexttime = dt_tm.Add(time.Duration(timeout+rand.Intn(600)) * 1e9)
	default:
		nexttime = dt_tm_00.Add(time.Duration((dlt*24*60+beg_tm*60)+dt_tm.Minute()) * 60 * 1e9)
	}

	time_to_sleep = nexttime.Sub(dt_tm)
	return
}

func Getrss() {
	for {
		resp, err := http.Get("http://94.45.209.196:1988/building.rss")
		// resp, err := http.Get("https://www.mskagency.ru/rss/building.rss")
		if err != nil {
			log.Println("get rss: ", err)
			panic(err)
		}
		defer resp.Body.Close()

		file, err := os.Create("./data/building.rss")
		if err != nil {
			log.Println("Unable to create file:", err)
			os.Exit(1)
		}
		io.Copy(file, resp.Body)
		defer file.Close()

		Check_news()

		nexttime, time_to_sleep := Time_to_sleep_f(40 * 60)
		log.Println("sleep, nexttime: ", nexttime)
		time.Sleep(time_to_sleep)
	}
}
