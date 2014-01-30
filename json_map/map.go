package json_map

import (
  "fmt"
  "os"
  "io/ioutil"
  "encoding/json"
  "strings"
  "github.com/tanema/go-socket.io"
)

type Map struct {
  Name        string            `json:"name"`
  Height      float32           `json:"height"`
  Width       float32           `json:"width"`
  TileHeight  float32           `json:"tileheight"`
  TileWidth   float32           `json:"tilewidth"`
  Layers      []*Layer          `json:"layers"`
  Orientation string            `json:"orientation"`
  TileSets    []*TileSet        `json:"tilesets"`
  Properties  map[string]string `json:"properties"`
  Player      *Sprite
  Players     map[string]*Sprite
  Npcs        map[string]*Sprite
  sio         *socketio.SocketIOServer
  Version     float32           `json:"version"`
}

func NewMap(map_path string, sio *socketio.SocketIOServer) *Map {
  file, e := ioutil.ReadFile(map_path)
  if e != nil {
    fmt.Printf("File error: %v\n", e)
    os.Exit(1)
  }
  var new_map Map
  json.Unmarshal(file, &new_map)

  new_map.Name = map_path[strings.LastIndex(map_path, "/")+1:len(map_path)]
  new_map.sio = sio
  new_map.Players = map[string]*Sprite{}
  new_map.Npcs = map[string]*Sprite{}

  new_map.initializeObjects()
  new_map.setupConnections()
  return &new_map
}

//this method normalizes all the objects position to tile position
//rather than abs coordinates
func (m *Map) initializeObjects(){
  for _, layer := range m.Layers {
    if layer.IsObjectGroup() {
      for _, sprite := range layer.Sprites {
        sprite.X = sprite.X / m.TileWidth
        sprite.Y = sprite.Y / m.TileHeight
        sprite.LayerName = layer.Name
        sprite.Map = m
        sprite.Layer = layer
        sprite.SetupSocket(m.sio)
        if sprite.IsNPC() {
          m.Npcs[sprite.Name] = sprite
        } else if sprite.IsPlayer() {
          m.Player = sprite
        }
      }
    }
  }
}

func (m *Map) setupConnections() {
  m.sio.Of(m.Name).On("connect", m.join)
  m.sio.Of(m.Name).On("disconnect", m.leave)
}

func (m *Map) join(ns *socketio.NameSpace){
  ns.Session.Values["map"] = m.Name

  new_player := m.Player.Clone()
  new_player.Id = ns.Id()
  if ns.Session.Values["x"] != nil && ns.Session.Values["y"] != nil {
    new_player.X = ns.Session.Values["x"].(float32)
    new_player.Y = ns.Session.Values["y"].(float32)
  }
  if ns.Session.Values["layer"] != nil {
    new_player.LayerName = ns.Session.Values["layer"].(string)
  }
  new_player.SetupSocket(m.sio)

  connected_data := struct {
    Player    *Sprite            `json:"player"`
    Players   map[string]*Sprite `json:"players"`
    Npcs      map[string]*Sprite `json:"npcs"`
  }{
    new_player,
    m.Players,
    m.Npcs,
  }

  //set the players initial data
  ns.Emit("player connected", connected_data);
  //tell everyone else about this player
  m.Players[new_player.Id] = new_player
  m.sio.In(m.Name).Except(ns).Broadcast("spawn", new_player);
}

func (m *Map) leave(ns *socketio.NameSpace){
  player := m.Players[ns.Id()]
  println(player.channel()+" disconnected")
  ns.Session.Values["x"] = player.X
  ns.Session.Values["y"] = player.Y
  ns.Session.Values["layer"] = player.LayerName
  ns.Session.Values["name"] = player.Name
  m.sio.In(player.channel()).Broadcast("kill");
  delete(m.Players, ns.Id())
}

func (m *Map) At(x, y float32) []MapObject {
  results := []MapObject{}
  for _, layer := range m.Layers {
    if layer.IsTileLayer() {
      tile_index := layer.Data[int(x + y * layer.Width)] - 1
      tile := m.TileSets[0].Tile(int64(tile_index))
      results = append(results, tile)
    } else if layer.IsObjectGroup() {
      for _, sprite := range layer.Sprites {
        if sprite.X == x && sprite.Y == y {
          results = append(results, sprite)
        }
      }
    }
  }
  return results
}
