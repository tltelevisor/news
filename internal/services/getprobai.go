package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	translator "github.com/Conight/go-googletrans"
)

type Role string

const (
	System    Role = "system"
	User      Role = "user"
	Assistant Role = "assistant"
)

type opai_err_mess struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}

type opai_err struct {
	Error opai_err_mess `json:"error"`
}

type Message_req struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type open_ai_req struct {
	Model    string        `json:"model"`
	Messages []Message_req `json:"messages"`
}

// Message represents an individual message in the chat
type Message struct {
	Role    Role    `json:"role"`
	Content string  `json:"content"`
	Refusal *string `json:"refusal"` // Refusal can be null, hence a pointer
}

// Choice represents each possible response generated in a chat completion
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	LogProbs     *int    `json:"logprobs"` // LogProbs can be null, hence a pointer
	FinishReason string  `json:"finish_reason"`
}

// TokenDetails provides specific details about token usage
type TokenDetails struct {
	CachedTokens             int `json:"cached_tokens"`
	ReasoningTokens          int `json:"reasoning_tokens"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
}

// UsageInfo provides details about tokens used in the conversation
type UsageInfo struct {
	PromptTokens            int          `json:"prompt_tokens"`
	CompletionTokens        int          `json:"completion_tokens"`
	TotalTokens             int          `json:"total_tokens"`
	PromptTokensDetails     TokenDetails `json:"prompt_tokens_details"`
	CompletionTokensDetails TokenDetails `json:"completion_tokens_details"`
}

// ChatCompletion represents the structure of a chat completion object
type ChatCompletion struct {
	ID                string    `json:"id"`
	Object            string    `json:"object"`
	Created           int64     `json:"created"`
	Model             string    `json:"model"`
	Choices           []Choice  `json:"choices"`
	Usage             UsageInfo `json:"usage"`
	SystemFingerprint string    `json:"system_fingerprint"`
}

func trns_ai(txt string) (res string) {
	t := translator.New()
	result, err := t.Translate(txt, "auto", "ru")
	if err != nil {
		panic(err)
	}
	res = result.Text
	return
}

func Find_prob_ai(text string) (prob float32, just string) {
	re := regexp.MustCompile(`\d{1}\.\d{1}|\d{1}\,\d{1}|\d{1}`)
	prob_text := re.FindString(text)
	log.Println("prob_text: ", prob_text)
	just_en := strings.TrimSpace(strings.Replace(text, prob_text, "", 1))
	just = trns_ai(just_en)
	prob_text = regexp.MustCompile(`,`).ReplaceAllString(prob_text, ".")
	prob64, err := strconv.ParseFloat(prob_text, 32)
	if err != nil {
		log.Println("ошибка преобразования во float: ", text)
	}
	prob = float32(prob64)
	if 0 > prob || prob > 1 {
		prob = 0.38
	}
	log.Println("prob: ", prob, "just: ", just)
	return
}

func Get_prob_open_ai(news string, product string) (prob float32, just string) {

	sys_prompt := fmt.Sprintf("Ты эксперт по продукту: %s", product)
	text_u := fmt.Sprintf("Оцени вероятность необходимости продукта для объекта из новости: %s В начале ответа число от 0 до 1, затем пояснение.", news)
	mess := []Message_req{{System, sys_prompt}, {User, text_u}}
	// model := "gpt-3.5-turbo-1106"
	model := "gpt-3.5-turbo-0125"
	data := open_ai_req{model, mess}
	json_data, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	url := "https://api.openai.com/v1/chat/completions"
	// url := "https://api.openai.com/v1/completions"

	log.Println("new Post: ", string(json_data)[:100])
	r := bytes.NewReader(json_data)

	req, err := http.NewRequest(http.MethodPost, url, r)
	if err != nil {
		log.Println("ошибка при формировании запроса: ", err)
	}
	req.Header.Add("Content-Type", "application/json")
	opapikey, exists := os.LookupEnv("OPENAI_API_KEY")
	if !exists {
		log.Println("Нет ключа OPENAI_API_KEY")
	}
	// fmt.Println("Bearer " + opapikey)
	req.Header.Add("Authorization", "Bearer "+opapikey)

	resp, err := http.DefaultClient.Do(req)
	// resp, err := http.Post(url, "application/json", r)
	if err != nil {
		panic(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Ошибка при чтении ответа:", err)
		return
	}
	var value ChatCompletion
	err = json.Unmarshal([]byte(body), &value)
	if err != nil {
		log.Println("Ошибка при Unmarshal body.response:", err)
	}

	if len(value.Choices) == 0 {
		log.Println("Ошибка: Choices пустой")
		log.Println(string(body))
		var value_err opai_err
		err = json.Unmarshal([]byte(body), &value_err)
		if err != nil {
			log.Println("Ошибка при Unmarshal error:", err)
		}
		log.Println(value_err.Error.Message)
	} else {
		resp_text := value.Choices[0].Message.Content
		log.Println(resp_text)
		// log.Println(resp_text[:100])
		prob, just = Find_prob_ai(resp_text)
	}
	return
}
