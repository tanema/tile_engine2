var Layer = require('./layer.js');

function Map(map_data){
  this.layers = [];
  this.player = null;
  this.objects = {};

  this.data = map_data;
  this.properties = map_data.properties || {};
  this.orientation = map_data.orientation;

  this._load_layer(map_data.layers);
}

Map.prototype._load_layer = function(layers){
  if(layers.length === 0){return;}
  var _map = this;
  new Layer(_map.data.layers[0], _map, function(layer){
    _map.layers.push(layer);
    layers.shift();
    _map._load_layer(layers);
  });
};

Map.prototype.at = function(x,y){
  var results = {tiles: [], objects: []};

  for(var i=0; i<this.layers.length; i++){
    var layer = this.layers[i];
    if(layer.is_tilelayer()){
      var tile = this.spritesheet.get(layer.data[(x + y * layer.width)]);
      if(tile){
        results.tiles.push(tile);
      }
    }else if(layer.is_objectgroup()){
      var object_name, object;
      for(object_name in layer.objects){
        object = layer.objects[object_name];
        if(object && object.x === x && object.y === y){
          results.objects.push(object);
        }
      }
    }
  }

  return results;
};

module.exports = Map;