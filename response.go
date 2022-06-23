package main

import (
	"fmt"
	"net/http"
)

func errorResponse(w http.ResponseWriter, msg string, errCode int) {
	newMsg := "[MS-MERCHANTS]: " + msg + ". Request, unsuccessful."
	//test := struct {
	//	Ok   bool     `json:"ok"`
	//	Msg  string   `json:"msg"`
	//	Data struct{} `json:"data"`
	//}{false, newMsg, struct{}{}}
	resp := fmt.Sprintf("{\"ok\": %v,\n\"msg\": \"%v\",\n\"data\": %v}", false, newMsg, struct{}{})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errCode)
	w.Write([]byte(resp))
	//json.NewEncoder(w).Encode(test)
}
