package delivery

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
	"word-search-in-files/pkg/searcher"
)

type SearchHandler struct {
	wordSearcher searcher.WordSearcher
	logger       *zap.SugaredLogger
}

func NewSearcherHandler(wordSearcher searcher.WordSearcher, logger *zap.SugaredLogger) *SearchHandler {
	return &SearchHandler{
		wordSearcher: wordSearcher,
		logger:       logger,
	}
}

func (sh *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		sh.logger.Errorf("keyword quety param is empty")
		err := writeResponse(w, http.StatusBadRequest, []byte(`{"message":"no keyword found in request"}`))
		if err != nil {
			sh.logger.Errorf("error in writing response: %s", err)
		}
		return
	}
	result := sh.wordSearcher.Search(keyword)
	if result == nil {
		sh.logger.Errorf("result is empty")
		err := writeResponse(w, http.StatusNotFound, []byte(`{"message":"keyword was not found in files"}`))
		if err != nil {
			sh.logger.Errorf("error in writing response: %s", err)
		}
		return
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		sh.logger.Errorf("error in JSON coding of result: %s", err)
		err = writeResponse(w, http.StatusInternalServerError, []byte(`{"message":"internal error"}`))
		if err != nil {
			sh.logger.Errorf("error in writing response: %s", err)
		}
		return
	}
	err = writeResponse(w, http.StatusOK, resultJSON)
	if err != nil {
		sh.logger.Errorf("error in writing response: %s", err)
	}
	return
}

func writeResponse(w http.ResponseWriter, status int, respBody []byte) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(status)
	_, err := w.Write(respBody)
	return err
}
