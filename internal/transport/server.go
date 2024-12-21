package transport

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"news/internal/services"
	"os"
	// "path/filepath"
)

func Check_oai_key(opapikey string) {

	// curl https://api.openai.com/v1/models \
	//   -H "Authorization: Bearer $OPENAI_API_KEY"

	type apitype struct {
		Isapikey bool `json:"isapikey"`
	}

	var isapitrue apitype = apitype{true}
	var isapifalse apitype = apitype{false}

	// fmt.Println(isapitrue)
	// svapi, _ := json.Marshal(isapitrue)
	// os.WriteFile("./ssession_id.json", svapi, 0644)
	// new, _ := os.ReadFile("./ssession_id.json")
	// var rdapi apitype
	// json.Unmarshal(new, &rdapi)
	// fmt.Println(rdapi)

	var url = "https://api.openai.com/v1/models"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println("ошибка при формировании запроса: ", err)
	}
	req.Header.Add("Authorization", "Bearer "+opapikey)

	resp, err := http.DefaultClient.Do(req)

	if resp.StatusCode != 200 {
		log.Println("Доступ запрещен", resp.Status)
		svapi, _ := json.Marshal(isapifalse)
		os.WriteFile("./data/isapi.json", svapi, 0644)
	} else {
		log.Println("Доступ разрешен", resp.Status)
		svapi, _ := json.Marshal(isapitrue)
		os.WriteFile("./data/isapi.json", svapi, 0644)
	}

}

func Server() {

	type apikey struct {
		Text string `json:"text"`
	}

	// session := "1111"
	// Установить признак доступности сервера Open AI api
	opapikey, exists := os.LookupEnv("OPENAI_API_KEY")
	if !exists {
		log.Println("Нет ключа OPENAI_API_KEY")
	}
	Check_oai_key(opapikey)

	fs := http.FileServer(http.Dir("./web/assets"))
	// Путь для файлов js-скриптов и стилей
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	fsd := http.FileServer(http.Dir("./data/"))
	// Путь для файлов данных
	http.Handle("/data/", http.StripPrefix("/data", fsd))

	// Данные для index.html
	// http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
	// 	file, err := os.ReadFile("./data/news_prod.json")
	// 	if err != nil {
	// 		file = []byte{}
	// 	}
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Header().Set("Access-Control-Allow-Origin", "*")
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Write(file)
	// })

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

		var error bool
		services.Opapikey, error = os.LookupEnv("OPENAI_API_KEY")
		if !error {
			log.Println("Нет ключа OPENAI_API_KEY")
		}

		services.Check_news_renew()

		file, err := os.ReadFile("./web/index.html")
		if err != nil {
			file = []byte{}
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write(file)
	})

	// Обработчик кнопки ввести API key и обновить
	http.HandleFunc("/rekey", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RemoteAddr, r.Body)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Ошибка при чтении body запроса apikey:", err)
			return
		}

		var value apikey
		err = json.Unmarshal(body, &value)
		if err != nil {
			log.Println("Ошибка при Unmarshal body.response apikey:", err)
		}
		services.Opapikey = value.Text
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

		file, err := os.ReadFile("./web/cv_ChakhovskyMA.pdf")
		if err != nil {
			file = []byte{}
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		w.Write(file)

	})

	log.Println("Server started on :8085")
	log.Fatal(http.ListenAndServe(":8085", nil))

	// defer os.Remove()
}
