package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	translator "github.com/Conight/go-googletrans"
)

func trns(txt string) (res string) {
	t := translator.New()
	result, err := t.Translate(txt, "auto", "ru")
	if err != nil {
		panic(err)
	}
	res = result.Text
	return
}

func Find_prob(text string) (prob float32, just string) {
	re := regexp.MustCompile(`\d{1}\,\d{1}|\d{1}`)
	prob_text := re.FindString(text)
	just_en := strings.TrimSpace(strings.Replace(text, prob_text, "", 1))
	just = trns(just_en)
	prob_text = regexp.MustCompile(`,`).ReplaceAllString(prob_text, ".")
	prob64, err := strconv.ParseFloat(prob_text, 32)
	if err != nil {
		log.Println("ошибка преобразования во float: ", text)
	}
	prob = float32(prob64)
	if 0 > prob || prob > 1 {
		prob = 0.33
	}
	log.Println("prob: ", prob, "just: ", just)
	return
}

func get_prob(news string, product string) (prob float32, just string) {

	type prgpt_resp_mess struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	type prgpt_resp_ch struct {
		Finish_reason string          `json:"finish_reason"`
		Delta         string          `json:"delta"`
		Message       prgpt_resp_mess `json:"message"`
		Sources       []string        `json:"sources"`
		Index         int             `json:"index"`
	}

	type prgpt_resp struct {
		Id      string          `json:"id"`
		Object  string          `json:"object"`
		Created int             `json:"created"`
		Model   string          `json:"model"`
		Choices []prgpt_resp_ch `json:"choices"`
	}

	type prgpt_req_mess struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	type prgpt_req struct {
		Messages    []prgpt_req_mess `json:"messages"`
		Stream      bool             `json:"stream"`
		Use_context bool             `json:"use_context"`
	}

	sys_prompt := fmt.Sprintf("Ты эксперт по продукту: %s", product)
	text_u := fmt.Sprintf("Оцени вероятность необходимости продукта для объекта из новости: %s Ответ дай числом от 0 до 1.", news)
	mess := []prgpt_req_mess{{"system", sys_prompt}, {"user", text_u}}
	// mess := []prgpt_req_mess{{"system", "one"}, {"user", "two"}}
	data := prgpt_req{mess, false, false}
	json_data, err := json.Marshal(data)
	if err != nil {
		log.Println("Ошибка в сборке запроса.", err)
		prob, just = 0.5, "Запрос был выполнен с недействительным ключом OPEN AI API key."
		return
	}

	log.Println("new Post: ", string(json_data)[:100])
	r := bytes.NewReader(json_data)
	resp, err := http.Post("http://5.180.174.86:8001/v1/chat/completions", "application/json", r)
	if err != nil {
		log.Println("Ошибка при отправке запроса серверу.", err)
		prob, just = 0.5, "Запрос был выполнен с недействительным ключом OPEN AI API key."
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Ошибка при чтении ответа сервера.", err)
		prob, just = 0.5, "Запрос был выполнен с недействительным ключом OPEN AI API key."
		return
	}
	var value prgpt_resp
	err = json.Unmarshal([]byte(body), &value)
	if err != nil {
		log.Println("Ошибка при разборе ответа сервера.", err)
		prob, just = 0.5, "Запрос был выполнен с недействительным ключом OPEN AI API key."
		return
	}

	resp_text := value.Choices[0].Message.Content
	var lnidx int
	if len(resp_text) > 100 {
		lnidx = 100
	} else {
		lnidx = len(resp_text)
	}
	log.Println(resp_text[:lnidx])
	prob, just = Find_prob(resp_text)
	return
}
