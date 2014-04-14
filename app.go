package main

import (
  "bytes"
  "encoding/json"
  "log"
  "net/http"
  "os"
  "sort"
  "strings"
)

var port = os.Getenv("PORT")
var endpoint = os.Getenv("SITUATION_ROOM_ENDPOINT")
var http_user = os.Getenv("SITUATION_ROOM_USERNAME")
var http_pass = os.Getenv("SITUATION_ROOM_PASSWORD")
var slackbot_endpoint = os.Getenv("SLACKBOT_REMOTE_ENDPOINT")
var secret = os.Getenv("SECRET_KEY")

type Rooms struct {
  Rooms map[string]Room
}
type Room struct {
  Available bool
}

func main() {
  log.Println("Notifier is starting up on :" + port)
  log.Println("Use Ctrl+C to stop")

  http.HandleFunc("/services/slackbot", slackbotHandler)
  http.ListenAndServe(":"+port, nil)
}

func slackbotHandler(w http.ResponseWriter, r *http.Request) {
  var channel = r.FormValue("channel_name")
  var request_secret = r.FormValue("secret")

  if (secret != request_secret) {
    http.Error(w, "Unauthenticated", 401)
    return
  }

  client := &http.Client{}

  req, err := http.NewRequest("GET", endpoint, nil)
  if err != nil {
    log.Fatal(err)
  }

  req.SetBasicAuth(http_user, http_pass)
  resp, err := client.Do(req)

  if err != nil {
    log.Println("Making request")
  	log.Fatal(err)
  }
  defer resp.Body.Close()

  var rooms Rooms
  err = json.NewDecoder(resp.Body).Decode(&rooms)
  if err != nil {
    log.Fatal(err)
  }

  availableRooms := make(sort.StringSlice,0)
  for key, value := range rooms.Rooms {
    if value.Available == true {
      availableRooms = append(availableRooms, key)
    }
  }

  availableRooms.Sort()

  var output string

  if len(availableRooms) == 0 {
    output = "No rooms currently available"
  } else {
    output = "Rooms currently available: \n"+ strings.Join(availableRooms, ", ")
  }
  buf := bytes.NewBufferString(output)

  dest := slackbot_endpoint +"&channel=%23" + channel
  resp, err = http.Post(dest, "application/x-www-form-urlencoded", buf)

  log.Println("Published message to "+ channel)
}
