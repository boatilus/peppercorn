export {}

declare global {
  interface Element {
    getAncestorByTagName(tag: String): Element;
    getFirstElementByClassName(className: String) : Element;
  }
}

// Retrieve the nearest ancestor that matches `tag`, returning `null` if it hits the <html> element.
Element.prototype.getAncestorByTagName = function(tag) {
  let e: Element = this;

  while (e = e.parentElement) {
    if (e.nodeName === 'HTML') return null;
    if (e.nodeName.toLowerCase() === tag) return e;
  }
};

Element.prototype.getFirstElementByClassName = function(className) {
  let e = this.getElementsByClassName(className);
  if (e.length === 0) {
    return null;
  }
  
  return e[0];
};