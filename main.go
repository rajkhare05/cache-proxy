package main

import (
    "encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
    "github.com/redis/go-redis/v9"
)

var args []string

func main() {

    localListeningPort, originHost := 0, ""

    // initialize redis client
    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })

    // get command line arguments
    args = os.Args

    // set port and origin host
    setHostAndPort(&localListeningPort, &originHost)

    address := fmt.Sprintf(":%d", localListeningPort)
    client := &http.Client{}

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        url := r.URL // request URL
        method := r.Method // request method

        // Only GET and HEAD methods are supported
        if method == "GET" || method == "HEAD" {

            // check if cached response exists
            responseExists, err := rdb.HExists(r.Context(), "response", url.String()).Result()
            if err != nil {
                log.Println("Error while checking cached response")
                log.Println(err)
            }

            // send the cached response if exists
            if responseExists {

                log.Println(method, url, "cache")

                // get cached response
                response, err := rdb.HGet(r.Context(), "response", url.String()).Result()
                if err != nil {
                    log.Println(err)
                }

                // add cached headers
                encodedHeaders, err := rdb.HGet(r.Context(), "headers", url.String()).Result()
                if err != nil {
                    log.Println(err)
                }

                var headers http.Header
                // parse the encoded headers
                json.Unmarshal([]byte(encodedHeaders), &headers)

                // add cached headers
                for k, v := range headers {
                    for _, vv := range v {
                        w.Header().Add(k, vv)
                    }
                }

                w.Header().Set("X-Cache", "HIT")
                // write cached response body
                fmt.Fprintf(w, "%s", response)
                return
            }
        }

        req := &http.Request{}
        serverURL := fmt.Sprintf("%s%s", originHost, url) // full server URL

        req, err := http.NewRequest(method, serverURL, nil)
        if err != nil {
            log.Println(err)
        }

        // add request headers
        for k, v := range r.Header {
            for _, vv := range v {
                req.Header.Add(k, vv)
            }
        }

        // request the origin server
        resp, err := client.Do(req)
        if err != nil {
            log.Println(err)
        }

        log.Println(method, url, resp.StatusCode)

        defer resp.Body.Close()
        // extract response body
        body, err := io.ReadAll(resp.Body)
        if err != nil {
            log.Println(err)
        }

        // add response headers
        for k, v := range resp.Header {
            for _, vv := range v {
                w.Header().Add(k, vv)
            }
        }

        if method == "GET" || method == "HEAD" {
            w.Header().Set("X-Cache", "MISS")
        }
        w.WriteHeader(resp.StatusCode)

        fmt.Fprintf(w, "%s", body)

        // save to cache
        err = rdb.HSet(r.Context(), "response", url.String(), string(body)).Err()
        if err != nil {
            log.Println(err)
        }

        headers, err := json.Marshal(resp.Header)
        if err != nil {
            log.Println(err)
        }

        err = rdb.HSet(r.Context(), "headers", url.String(), headers).Err()
        if err != nil {
            log.Println(err)
        }

    })

    log.Print("[Cache proxy started on ", localListeningPort, "]")
    err := http.ListenAndServe(address, nil)

    if err != nil {
        log.Fatal(err)
    }
}

func setHostAndPort(localListeningPort *int, originHost *string) {
    argsLength := len(args)
    err := error(nil)

    for i := range args {
        if args[i] == "--port" && i + 1 < argsLength {
            *localListeningPort, err = strconv.Atoi(args[i + 1])

            if err != nil {
                fmt.Println("Error while parsing local listening port")
                portError()

            } else if *localListeningPort < 21 || *localListeningPort > 65535 {
                portError()
            }

        } else if args[i] == "--origin" {
            if i + 1 < argsLength {
                *originHost = args[i + 1]

            } else {
                fmt.Println("Origin host is required")
                originError()
            }
        }
    }
}

func portError() {
    fmt.Println("Missing local listening port [Port range 21 - 65535] \nUsage: caching-proxy --port <port> --origin <host>")
    os.Exit(1)
}

func originError() {
    fmt.Println("Missing origin host \nUsage: caching-proxy --port <port> --origin <host>")
    os.Exit(1)
}

