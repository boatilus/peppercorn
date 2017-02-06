export {}

declare global {
  interface Event {
    isModified(): boolean;
  }
}

Event.prototype.isModified = function() {
  return this.ctrlKey || this.getModifierState("Meta");
};