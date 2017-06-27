package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "strconv"
)

type Player struct {
  id int
  firstname string
  secondname string
  price int
  totalpoints int
  gamepoints []int
  fivegamepoints int
  tengamepoints int
  team int
  position int
}

func main() {
  //https://fantasy.premierleague.com/drf/bootstrap-static
  //https://fantasy.premierleague.com/drf/elements/
  //https://fantasy.premierleague.com/drf/element-summary/1

  playerdata := playerdata()
  pdlength := len(playerdata)
  gameweek := gameweek()
  fmt.Printf("Gameweek: %d\n", gameweek)
  var players []Player

  for k, p := range playerdata {
    p := p.(map[string]interface{})
    pp := player(p, gameweek)
    players = append(players, pp)
    fmt.Printf("\rPlayer: %d/%d", k+1, pdlength)
  }

  export(players)
}

//Takes the basic data map for a player and the gameweek, and returns the Player struct version
func player(p map[string]interface{}, gw int) Player {
  resp, err := http.Get(string(strconv.AppendFloat([]byte("https://fantasy.premierleague.com/drf/element-summary/"), p["id"].(float64), 'f', 0, 32)))
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    panic(err)
  }

  var pdetails map[string]interface{}
  if err := json.Unmarshal(body, &pdetails); err != nil {
    panic(err)
  }

  var gamepoints []int
  for i := 0; i < gw; i++ {
    gamepoints = append(gamepoints, 0)
  }

  var roundsplayed []int
  for _, gp := range pdetails["history"].([]interface{}) {
    gp := gp.(map[string]interface{})
    if len(roundsplayed) == 0 {
      roundsplayed = append(roundsplayed, int(gp["round"].(float64)))
    }
    if roundsplayed[len(roundsplayed)-1] != int(gp["round"].(float64)) {
      roundsplayed = append(roundsplayed, int(gp["round"].(float64)))
    }
    gamepoints[int(gp["round"].(float64))-1] += int(gp["total_points"].(float64))
  }

  fivegamepoints := 0
  tengamepoints := 0
  for i := 0; i < 5; i++ {
    if i < len(roundsplayed) {
      fivegamepoints += gamepoints[roundsplayed[len(roundsplayed)-1-i]-1]
    }
  }
  tengamepoints += fivegamepoints
  for i := 5; i < 10; i++ {
    if i < len(roundsplayed) {
      tengamepoints += gamepoints[roundsplayed[len(roundsplayed)-1-i]-1]
    }
  }

  return Player{
    id: int(p["id"].(float64)), //int
    firstname: p["first_name"].(string), //string
    secondname: p["second_name"].(string), //string
    price: int(p["now_cost"].(float64)), //int
    totalpoints: int(p["total_points"].(float64)), //int
    gamepoints: gamepoints,
    fivegamepoints: fivegamepoints,
    tengamepoints: tengamepoints,
    team: int(p["team"].(float64)), //int
    position: int(p["element_type"].(float64)), //int
  }
}

//Returns last gameweek played
func gameweek() int {
  gameweek := 0

  resp, err := http.Get("https://fantasy.premierleague.com/drf/events")
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    panic(err)
  }

  var gameweeks []interface{}
  if err := json.Unmarshal(body, &gameweeks); err != nil {
    panic(err)
  }

  for _, gw := range gameweeks {
    gw := gw.(map[string]interface{})

    if gw["finished"] == true {
      gameweek += 1
    } else {
      return gameweek // gameweek + 1 if you want latest mid-gameweek stats included
    }
  }

  return gameweek
}

//Returns the array of all players and basic data
func playerdata() []interface{} {
  resp, err := http.Get("https://fantasy.premierleague.com/drf/elements")
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    panic(err)
  }

  var playerdata []interface{}
  if err := json.Unmarshal(body, &playerdata); err != nil {
    panic(err)
  }

  return playerdata
}

func export(ps []Player) {
  s := string([]byte{0xEF,0xBB,0xBF}) //UTF-8 BOM
  s += "ID,Second Name,First Name,Team,Position,Price,"
  for i := 1; i <= len(ps[0].gamepoints); i++ {
    s += "GW" + strconv.Itoa(i) + ","
  }
  s += "Total Points,Ten Game Points,Five Game Points\n"

  for _, p := range ps {
    s = s + strconv.Itoa(p.id) + ","
    s = s + p.secondname + ","
    s = s + p.firstname + ","
    s = s + strconv.Itoa(p.team) + ","
    s = s + strconv.Itoa(p.position) + ","
    s = s + strconv.Itoa(p.price) + ","
    for _, gwp := range p.gamepoints {
      s = s + strconv.Itoa(gwp) + ","
    }
    s = s + strconv.Itoa(p.totalpoints) + ","
    s = s + strconv.Itoa(p.tengamepoints) + ","
    s = s + strconv.Itoa(p.fivegamepoints) + "\n"
  }

  err := ioutil.WriteFile("data.csv", []byte(s), 0644)
  if err != nil {
    panic(err)
  }
}
