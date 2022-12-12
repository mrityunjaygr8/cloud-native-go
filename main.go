package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	patterns "github.com/mrityunjaygr8/cloud-native-go/patterns"
)

func failAfter(threshold int) patterns.Circuit {
	count := 0

	return func(ctx context.Context) (string, error) {
		count++

		if count > threshold {
			return "", errors.New("INTENTIONAL FAIL!")
		}

		return "Success", nil
	}
}

type jsonObj map[string]string

func writeToJSON(w http.ResponseWriter, message jsonObj) {
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(message)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.Write(jsonResp)
	return

}

func main() {
	fmt.Println("Starting server at port 9000")
	circ := failAfter(5)
	breaker := patterns.Breaker(circ, 1)
	ctx := context.Background()

	http.HandleFunc("/threshold", func(w http.ResponseWriter, r *http.Request) {

		res, err := breaker(ctx)
		resp := make(jsonObj)
		if err != nil {
			resp["error"] = err.Error()
			writeToJSON(w, resp)
			return
		}

		resp["body"] = res
		writeToJSON(w, resp)
		return
	})
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}
