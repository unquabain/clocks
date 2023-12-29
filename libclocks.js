
const xmlns = 'http://www.w3.org/2000/svg';
let svg = document.getElementsByTagName('svg')[0];
let clockFace;

function makeA(what)
{
  return document.createElementNS(xmlns, what);
}

function add(el)
{
  clockFace.appendChild(el);
  return el;
}

function Path()
{
  function xy(x, y)
  {
    return x + ',' + y;
  }

  function _(...list)
  {
    return list.join(' ');
  }

  this.instructions = [];
  this.lastXY = [0, 0];

  this.updateXY = function(x, y)
  {
    this.lastXY = [x, y];
  }

  this.updateDXY = function(dx, dy)
  {
    let x = this.lastXY[0] + dx;
    let y = this.lastXY[1] + dy;
    this.updateXY(x, y);
  }

  this.m = function(dx, dy) {
    this.instructions.push('m' + xy(dx,dy));
    this.updateDXY(dx, dy);
    return this;
  }

  this.M = function(x, y) {
    this.instructions.push('M' + xy(x,y));
    this.updateXY(x, y);
    return this;
  }

  this.polar_m = function(r, theta) {
    let dx = Math.cos(theta) * r;
    let dy = Math.sin(theta) * r;
    return this.m(dx, dy);
  }

  this.l = function(dx, dy) {
    this.instructions.push('l' + xy(dx,dy));
    this.updateDXY(dx, dy);
    return this;
  }

  this.L = function(x, y) {
    this.instructions.push('L' + xy(x,y));
    this.updateXY(x, y);
    return this;
  }

  this.polar_l = function(r, theta) {
    let dx = Math.cos(theta) * r;
    let dy = Math.sin(theta) * r;
    return this.l(dx, dy);
  }

  this.z = function() { this.instructions.push('z'); return this }
  this.Z = function() { this.instructions.push('Z'); return this; }

  this.h = function(dx) {
    this.instructions.push('h' + dx);
    this.lastXY = [ this.lastXY[0] + dx, this.lastXY[1] ];
    return this;
  }
  this.H = function(x) {
    this.instructions.push('H' + x);
    this.lastXY = [ x, this.lastXY[1] ];
    return this;
  }

  this.v = function(dy) {
    this.instructions.push('v' + dy);
    this.lastXY = [ this.lastXY[0], this.lastXY[1] + dy ];
    return this;
  }

  this.V = function(y) {
    this.instructions.push('V' + y);
    this.lastXY = [ this.lastXY[0], y ];
    return this;
  }

  this.c = function(dx1, dy1, dx2, dy2, dx, dx) {
    this.instructions.push(
      'c' + _(xy(dx1,dy1), xy(dx2,dy2), xy(dx,dy))
    );
    this.updateDXY(dx, dy);
    return this;
  }

  this.C = function(x1, y1, x2, y2, x, y) {
    this.instructions.push(
      'C' + _(xy(x1,y1), xy(x2,y2), xy(x,y))
    );
    this.updateXY(x, y);
    return this;
  }

  this.s = function(dx2, dy2, dx, dy) {
    this.instructions.push(
      's' + _(xy(dx2, dy2), xy(dx, dy))
    );
    this.updateDXY(dx, dy);
    return this;
  }

  this.S = function(x2, y2, x, y) {
    this.instructions.push(
      'S' + _(xy(x2, y2), xy(x, y))
    );
    this.updateXY(x, y);
    return this;
  }

  this.q = function(dx2, dy2, dx, dy) {
    this.instructions.push(
      'q' + _(xy(dx2, dy2), xy(dx, dy))
    );
    this.updateDXY(dx, dy);
    return this;
  }

  this.Q = function(x2, y2, x, y) {
    this.instructions.push(
      'Q' + _(xy(x2, y2), xy(x, y))
    );
    this.updateXY(x, y);
    return this;
  }

  this.t = function(dx, dy) {
    this.instructions.push(
      't' + xy(dx, dy)
    );
    this.updateDXY(dx, dy);
    return this;
  }

  this.T = function(x, y) {
    this.instructions.push(
      'T' + xy(x, y)
    );
    this.updateXY(x, y);
    return this;
  }

  this.a = function(rx, ry, rot, arc, sweep, dx, dy) {
    this.instructions.push(
      'a' + _(
        xy(rx, ry),
        rot, arc, sweep,
        xy(dx, dy)
      )
    );
    this.updateDXY(dx, dy);
    return this;
  }

  this.A = function(rx, ry, rot, arc, sweep, x, y) {
    this.instructions.push(
      'A' + _(
        xy(rx, ry),
        rot, arc, sweep,
        xy(x, y)
      )
    );
    this.updateXY(x, y);
    return this;
  }

  this.toString = function()
  {
    return this.instructions.join("\n");
  }

  this.toEl = function()
  {
    let path = makeA('path');
    path.setAttribute('d', this.toString());
    return path;
  }

  return this;
}

function nodeMapToArray(attributes)
{
  let aa = {};
  for (var key in attributes) {
    if (! attributes.hasOwnProperty(key)) continue;
    let att = attributes[key];
    aa[att.name] = att.value;
  }
  return aa;
}

function write(text, x, y, size)
{
  let tnode = document.createTextNode(text);
  let tel = makeA('text');
  tel.appendChild(tnode);
  setAttributes(tel, {
    'font-size': size,
    x: x,
    y: y,
    'text-anchor': 'middle',
    'alignment-baseline': 'middle'
  });
  return tel;
}

function setAttributes(el, attributes) {
  for (var key in attributes) {
    if (! attributes.hasOwnProperty(key)) continue;
    let val = attributes[key];
    el.setAttribute(key, val);
  }
  return el;
}

function currentTime()
{
  let now = new Date();
  let o = {};
  o.sec = {};
  o.sec.i = now.getSeconds();
  o.sec.f = o.sec.i + (now.getMilliseconds() / 1000.0);
  o.sec.p = o.sec.f / 60.0;
  o.sec.r = o.sec.p * 2 * Math.PI;
  o.sec.d = o.sec.p * 360;

  o.min = {};
  o.min.i = now.getMinutes();
  o.min.f = o.min.i + (o.sec.p);
  o.min.p = o.min.f / 60.0;
  o.min.r = o.min.f * 2 * Math.PI;
  o.min.d = o.min.f * 360;

  o.hr = {}
  o.hr.i = now.getHours();
  o.hr.f = o.hr.i + o.min.p;
  o.hr.p12 = (o.hr.f % 12) / 12.0;
  o.hr.p24 = o.hr.f / 24.0;
  o.hr.r12 = o.hr.p12 * 2 * Math.PI;
  o.hr.r24 = o.hr.p24 * 2 * Math.PI;
  o.hr.d12 = o.hr.p12 * 360;
  o.hr.d24 = o.hr.p24 * 360;

  return o;
}

/* Chrome removed the ability to style elements referred to
 * by a <use> element, because they have a better plan to replace
 * it ... which they haven't implemented yet. (ノಠ益ಠ)ノ彡┻━┻
 */
function realizeUses()
{
  let uses = document.getElementsByTagName('use');
  while (uses.length > 0) {
    let use = uses.item(0);
    let href = use.getAttribute('href');
    let id = href.split('#')[1];
    let el = document.getElementById(id);
    let clone = el.cloneNode(true);
    let attributes = nodeMapToArray(use.attributes);
    delete attributes.href;
    clone.removeAttribute('id');
    setAttributes(clone, attributes);
    use.parentElement.insertBefore(clone, use);
    use.remove();
  };
}

function use(id, attributes)
{
  let source = document.getElementById(id);
  let clone = source.cloneNode(true);
  return setAttributes(clone, attributes);
}

function init()
{
  clockFace = svg.appendChild(makeA('g'));
  realizeUses();
  drawFace();
  tick();
  setInterval(tick, 1000/FRAME_RATE);
}
