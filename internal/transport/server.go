package transport

import (
	"log"
	"net/http"
	"news/internal/services"
	"os"
	// "path/filepath"
)

func Server() {
	fs := http.FileServer(http.Dir("./web/assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Данные для index.html
	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.ReadFile("./data/news_prod.json")
		if err != nil {
			file = []byte{}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write(file)
	})

	// Главный обработчик для отображения index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RemoteAddr)
		file, err := os.ReadFile("./web/index.html")
		if err != nil {
			file = []byte{}
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write(file)
	})

	// Обработчик кнопки обновить
	http.HandleFunc("/renew", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RemoteAddr)

		services.Check_news_renew()
		file, err := os.ReadFile("./web/index.html")
		if err != nil {
			file = []byte{}
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write(file)
	})

	// Обработчик для отображения резюме
	http.HandleFunc("/cv", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RemoteAddr)

		file, err := os.ReadFile("./web/cv_ChakhovskyMA_18.11.24.pdf")
		if err != nil {
			file = []byte{}
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		w.Write(file)

	})

	log.Println("Server started on :8083")
	log.Fatal(http.ListenAndServe(":8083", nil))

}
