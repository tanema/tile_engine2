package main

import (
	"net/http"
  "log"
  "github.com/tanema/go-socket.io"
  "github.com/tanema/grouter/json_map"
)

var players map[string]*json_map.Player
var maps map[string]*json_map.Map

func main() {
  sio := socketio.NewSocketIOServer(&socketio.Config{})
  sio.On("connect", func(ns *socketio.NameSpace){
    log.Println("Connected: ", ns.Id())
  })
  sio.On("disconnect", func(ns *socketio.NameSpace){
    player := players[ns.Id()]
    ns.Session.Values["x"] = player.X
    ns.Session.Values["y"] = player.Y
    ns.Session.Values["layer"] = player.LayerName
    ns.Session.Values["name"] = player.Name
    log.Println("Disconnected: ", ns.Id())
    sio.Broadcast("kill player", ns.Id());
    delete(players, ns.Id())
  })
  sio.On("player move", func(ns *socketio.NameSpace, to_x, to_y int){
    //TODO validate move
    player := players[ns.Id()]
    player.X = float32(to_x)
    player.Y = float32(to_y)
    sio.Except(ns).Broadcast("actor move", ns.Id(), to_x, to_y);
  })
  sio.On("player change layer", func(ns *socketio.NameSpace, layer string){
    log.Println("player change layer: ", ns.Id())
    player := players[ns.Id()]
    player.LayerName = layer
    sio.Except(ns).Broadcast("actor change layer", ns.Id(), layer);
  })
  sio.On("player teleport", func(ns *socketio.NameSpace, to_x, to_y int){
    //TODO validate move
    player := players[ns.Id()]
    player.X = float32(to_x)
    player.Y = float32(to_y)
    sio.Except(ns).Broadcast("actor teleport", ns.Id(), to_x, to_y);
  })
  sio.On("join map", func(ns *socketio.NameSpace, map_name, player_name string){
    log.Println("join map")
    ns.Session.Values["map"] = map_name
    ns.Session.Values["name"] = player_name

    new_player := maps[map_name].Player.ShallowClone()
    new_player.Id = ns.Id()
    new_player.Name = player_name
    if ns.Session.Values["x"] != nil && ns.Session.Values["y"] != nil {
      new_player.X = ns.Session.Values["x"].(float32)
      new_player.Y = ns.Session.Values["y"].(float32)
    }
    if ns.Session.Values["layer"] != nil {
      new_player.LayerName = ns.Session.Values["layer"].(string)
    }

    connected_data := struct {
      Player    *json_map.Player            `json:"player"`
      Players   map[string]*json_map.Player `json:"players"`
      Npcs      map[string]*json_map.Sprite `json:"npcs"`
    }{
      new_player, players, map[string]*json_map.Sprite{},
    }

    //set the players initial data
    ns.Emit("player connected", connected_data);
    //tell everyone else about this player
    players[new_player.Id] = new_player
    sio.Except(ns).Broadcast("spawn player", new_player);
  })
  sio.On("set name", func(ns *socketio.NameSpace, name string){
    println("set name")
    ns.Session.Values["name"] = name
    sio.Broadcast("change name", ns.Id(), name)
  })

  players = map[string]*json_map.Player{}
  maps = map[string]*json_map.Map{
    "map0.json": json_map.NewMap("public/maps/map0.json"),
  }

  sio.Handle("/", http.FileServer(http.Dir("./public/")))

	println("listening on port 3000")
  log.Fatal(http.ListenAndServe(":3000", sio))
}