package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mrityunjaygr8/cloud-native-go/patterns"
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

// func
func failAfterEffector(threshold int) patterns.Effector {
	count := 0

	return func(ctx context.Context) (string, error) {
		count++

		if count < threshold {
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
	circEffector := failAfterEffector(3)
	breaker := patterns.Breaker(circ, 1)
	debounce_first := patterns.DebounceFirst(circ, time.Second)
	debounce_last := patterns.DebounceLast(circ, time.Second)
	retry := patterns.Retry(circEffector, 1, time.Second)
	throttle := patterns.Throttle(circEffector, 1, 1, time.Second)
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

	http.HandleFunc("/debounce-first", func(w http.ResponseWriter, r *http.Request) {
		res, err := debounce_first(ctx)
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
	http.HandleFunc("/debounce-last", func(w http.ResponseWriter, r *http.Request) {
		res, err := debounce_last(ctx)
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
	http.HandleFunc("/retry", func(w http.ResponseWriter, r *http.Request) {
		res, err := retry(ctx)
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
	http.HandleFunc("/throttle", func(w http.ResponseWriter, r *http.Request) {
		res, err := throttle(ctx)
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
