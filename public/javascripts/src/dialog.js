//TODO stop the player while talking and timeout after the end of tlaking
function Dialog(){
  this.dialog_height = 25,
  this.padding_left = 10,
  this.padding_bottom = 5;
  this.is_talking = false;
  this.script = [];
  this.dialog_open_length = 1000;
  this.can_close = false;
}

Dialog.prototype.draw = function(ctx){
  //nothing to see here
  if(!this.is_talking){return;}

  // first give text a background
  ctx.fillStyle = '#dedede';
  ctx.fillRect(0, (ctx.canvas.height - this.dialog_height), ctx.canvas.width, this.dialog_height);
  // then draw text
  ctx.font = '20px pokemon';
  ctx.fillStyle = 'black';
  ctx.fillText(this.script[0], this.padding_left, (ctx.canvas.height - this.padding_bottom));
};

Dialog.prototype.say = function(script){
  if(this.just_closed){return;}

  if(typeof script === "object"){ //pass in an array of things to say
    this.script = script;
  }else if(typeof script === "string"){
    this.script = [script];
  }
  this.is_talking = true;
  this._after_new_dialog();
};

Dialog.prototype.next = function(){
  //wait until you can close, this prevents the fast button repeat
  if(!this.can_close){return;}

  if(this.script.length > 1){
    this.script.shift();
  }else{
    this.is_talking = false;
    var _this = this;
    this.just_closed = true;
    setTimeout(function(){
      _this.just_closed = false;
    }, this.dialog_open_length);
  }

  this._after_new_dialog();
};

Dialog.prototype._after_new_dialog = function(){
  $(document).trigger("redraw");
  this.can_close = false;
  var _this = this;
  setTimeout(function(){
    _this.can_close = true;
  }, this.dialog_open_length);
};