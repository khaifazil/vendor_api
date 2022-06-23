package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

func genApiKey() (string, error) {
	b := make([]byte, 20)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	hash.Write(b)

	apiKey := hex.EncodeToString(hash.Sum(nil))

	return apiKey, nil
}

func getNewApiKey(w http.ResponseWriter, r *http.Request) {
	if apiKey, err := genApiKey(); err != nil {
		errorResponse(w, "unable to generate apikey", 500)
		ErrorLogger.Println("unable to generate apikey", err)
		return
	} else {
		db := openDatabase()
		defer closeDatabase(db)

		_, err := db.Exec("INSERT INTO api_key(API_Keys) VALUE (?)", apiKey)
		if err != nil {
			errorResponse(w, "unable query database", 500)
			ErrorLogger.Println("unable query database.", err)
			return
		}

		type apikey struct {
			Apikey string `json:"apikey"`
		}

		reply := struct {
			Ok   bool   `json:"ok"`
			Msg  string `json:"msg"`
			Data apikey `json:"data"`
		}{
			Ok:   true,
			Msg:  "[MS-MERCHANTS]: ApiKey generation, successful",
			Data: apikey{apiKey},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(202)
		json.NewEncoder(w).Encode(reply)
	}
}

func validateAPIKey(r *http.Request) bool {
	apikey := r.Header.Get("key")

	db := openDatabase()
	defer closeDatabase(db)

	err := db.QueryRow("SELECT API_Keys FROM api_key WHERE API_Keys = ?", apikey).Scan(&apikey)
	if err != nil {
		if err != sql.ErrNoRows {
			return false
		}
		return false
	}
	return true
}
