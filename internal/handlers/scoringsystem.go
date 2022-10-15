package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"gofermart/internal/compression"
	"gofermart/internal/constants"
	"gofermart/internal/postgresql"
)

func GetScoringSystem(number string) (scoringSystem *postgresql.ScoringSystem, httpStatus int) {

	scoringSystem = new(postgresql.ScoringSystem)

	addressPost := fmt.Sprintf("http://%s/api/orders/%s", "localhost:8000", number)
	req, err := http.NewRequest("GET", addressPost, strings.NewReader(""))
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	defer req.Body.Close()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}
	defer resp.Body.Close()

	varsAnswer := mux.Vars(req)
	fmt.Println(varsAnswer)

	bodyJSON := resp.Body

	contentEncoding := resp.Header.Get("Content-Encoding")
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			constants.Logger.ErrorLog(err)
			return scoringSystem, http.StatusInternalServerError
		}

		arrBody, err := compression.Decompress(bytBody)
		//if err != nil {
		//	constants.Logger.ErrorLog(err)
		//	http.Error(w, "Ошибка распаковки", http.StatusInternalServerError)
		//	return
		//}

		//bodyJSON = bytes.NewReader(arrBody)
		fmt.Println(arrBody)
	}

	if err := json.NewDecoder(bodyJSON).Decode(&scoringSystem); err != nil {
		constants.Logger.InfoLog(fmt.Sprintf("$$ 3 %s", err.Error()))
		return scoringSystem, http.StatusInternalServerError
	}

	return scoringSystem, http.StatusOK
}
