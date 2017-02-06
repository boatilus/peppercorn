/// <reference path="typings/modules/markdown-it/index.d.ts" />

import 'element';
import 'event';
import * as icon from './icon';
import * as markdownit from 'markdown-it';

const md = new markdownit({
  linkify: true,
  typographer: true
});

const returnKey  = 13;
const shiftKey   = 16;
const escapeKey  = 27;
const leftArrow  = 37;
const upArrow    = 38;
const rightArrow = 39;
const downArrow  = 40;

let isAdmin     = false;
let currentUser = '';

let prev   = null;
let next   = null;
let bottom = null;

// Given a potentially multi-line string of text, return a version of that text with any
// Markdown blockquotes removed.
const stripQuotes = function(text) {
  let lines = text.split(/\r?\n/);
  let newlines = [];

  for (let i = 0; i < lines.length; i++) {
    let line = lines[i].trim();

    // We know this line is blockquotes if it begins with `>`.
    if (line.charAt(0) !== '>') {
      newlines.push(line);
    }
  }
  
  if (newlines[0] === '') {
    newlines.shift();
  }

  return newlines.join(`\r\n`);
};

// Given a potentially multi-line string of text, return a version of that text with a `>`
// prepended to each line for a Markdown blockquote.
const quote = function(text) {
  let lines = text.split(/\r?\n/);
  let newlines = [`> **User**:`];

  for (let i = 0; i < lines.length; i++) {
    let line = lines[i].trim();
    
    newlines.push('> ' + line);
  }

  newlines.push(`\r\n`);
  
  return newlines.join(`\r\n`);
}

const handleKeyDownEvents = function(event) {
  const nodeName = document.activeElement.nodeName;
  const isInputFocused = nodeName === 'TEXTAREA' || nodeName === 'INPUT';

  // We want users to be able to hit Esc and get out of the reply field so they can get back to
  // using shortcuts for everything else.
  if (isInputFocused) {
    if (event.keyCode === escapeKey) {
      if (typeof document.activeElement === 'HTMLElement') {
        let ae = <HTMLElement>document.activeElement;
        ae.blur();
      }
    } else {
      return false;
    }
  }
  
  // We'll want to prevent all the following if a textfield or input is focused, as we don't want
  // to cause problems with a user's text entry.
  switch (event.keyCode) {
  case shiftKey:
    document.getElementById('page-prev').style.visibility = 'visible';
    document.getElementById('page-next').style.visibility = 'visible';
    return;
  case leftArrow:
    if (event.shiftKey && (prev !== null)) window.location.href = prev.getAttribute('href');
    return;
  case rightArrow:
    if (event.shiftKey && (next !== null)) window.location.href = next.getAttribute('href');
    return;
  case upArrow:
    if (event.shiftKey) window.scrollTo(0, 0);
    return;
  case downArrow:
    if (event.shiftKey) {
      window.scrollTo(0, document.body.clientHeight);
      bottom.focus();
    }
  }
};

const handleKeyUpEvents = function(event) {
  if (event.keyCode === shiftKey) {
    document.getElementById('page-prev').style.visibility = 'hidden';
    document.getElementById('page-next').style.visibility = 'hidden';
    return;
  }
};

// Accepts an <article> element and returns its content as a trimmed string.
const getTrimmedContent = function(articleElem) {
  const content = articleElem.getFirstElementByClassName('article-content');
  if (content === null) {
    console.error('handleReplyClick: no element found for this post with article-content');

    return false;
  }

  return content.textContent.trim();
};

const handleReplyClick = function(event) {
  const article = this.getAncestorByTagName('article');
  if (article === null) {
    console.error('Could not find article ancestor');
    return false;
  }

  const trimmedContent = getTrimmedContent(article);
  const strippedAndQuoted = quote(stripQuotes(trimmedContent));

  bottom.value = strippedAndQuoted;
  bottom.focus();
};

// handleEditClick is the handler called for the Edit button `click` event.
const handleEditClick = function(event) {
  const displayViewState = function() {
    editable.remove();
    editable = null;

    rendered.style.display = 'block';
  };
  
  // We'll create a textarea element filled with the post's Markdown comment right within the
  // post, then temporarily hide the rendered content. On Ctrl+Enter or ⌘+Enter, we'll PATCH
  // the post with the new content and re-render with revised content on success. If the user hits
  // Esc, we'll simply remove the textarea element and show the existing rendered content. 
  const article = this.getAncestorByTagName('article');
  if (article === null) {
    console.error('handleEditClick: could not find article ancestor');
    return false;
  }

  let content = article.getFirstElementByClassName('article-content');
  if (content === null) {
    console.error('handleEditClick: no element found for this post with article-content');
    return false;
  }

  let rendered = article.getFirstElementByClassName('article-rendered');
  if (rendered === null) {
    console.error('handleEditClick: no element found for this post with article-rendered');
    return false;
  }
  
  let editable = document.createElement('textarea');
  editable.className = 'article-editable';
  editable.value     = content.textContent;
  editable.rows      = 4;

  const handleKeydown = function(e) {
    if (e.keyCode === escapeKey) {
      displayViewState();
      return;
    }

    if (e.isModified() && (e.keyCode === returnKey)) {
      const val = editable.value;

      // If no changes to the post's content, skip the entire submission process and just swap
      // back to the post's normal view state.
      if (val === content.innerHTML) {
        console.log(`handleEditClick: no change to content of "${article.id}"`);

        displayViewState();
        return;
      }

      let xhr = new XMLHttpRequest();
      xhr.open('PATCH', `/posts/${article.id}`, true);
      xhr.setRequestHeader('Content-type', 'application/json');
      xhr.addEventListener('loadstart', function() {
        console.time('post-edit');
        console.log(`handleEditClick: sending PATCH request for "${article.id}"..`);
      });
      xhr.addEventListener('loadend', function() {
        console.timeEnd('post-edit');
        console.log(`handleEditClick: response for PATCH request received for "${article.id}"`);

        // Store text in `val` because we'll remove the textarea from the DOM.
        const val = editable.value;

        // Keep the new Markdown in `article-content`, and render it to `article-rendered`.
        content.innerHTML  = val;
        rendered.innerHTML = md.render(val);
        displayViewState();
      });
      xhr.addEventListener('timeout', function() {
        console.timeEnd('post-edit');
        console.error("handleEditClick: PATCH request timed out");
      });
      xhr.send(JSON.stringify({ content: editable.value }));
    }
  }

  editable.addEventListener('keydown', handleKeydown);

  article.appendChild(editable);
  editable.focus();

  rendered.style.display = 'none';
};

const handleDeleteClick = function(event) {
  const article = this.getAncestorByTagName('article');
  if (article === null) {
    console.error('Could not find article ancestor');
    return false;
  }

  if (window.confirm('Are you sure you want to delete this post?')) {
    window.location.href = `/posts/${article.id}/delete`;
  }
};

document.addEventListener('DOMContentLoaded', function() {
  isAdmin     = (document.body.dataset['isAdmin'] === 'true');
  currentUser = document.body.dataset['currentUser'];

  prev   = document.getElementById('nav-previous');
  next   = document.getElementById('nav-next');
  bottom = document.getElementById('bottom');

  let prevArrow = document.createElement('div');
  prevArrow.id = 'page-prev';
  if (prev !== null) {
    prevArrow.className = 'page-prev-enabled';
  }

  let nextArrow = document.createElement('div');
  nextArrow.id = 'page-next';
  if (next !== null) {
    nextArrow.className = 'page-next-enabled';
  }

  let arrowFragment = document.createDocumentFragment();
  arrowFragment.appendChild(prevArrow);
  arrowFragment.appendChild(nextArrow);

  document.body.appendChild(arrowFragment);

  window.addEventListener('keydown', handleKeyDownEvents);
  window.addEventListener('keyup', handleKeyUpEvents);

  console.time('DOM_begin');
  
  let posts = document.getElementsByTagName('article');

  for (let i = 0; i < posts.length; i++) {
    const thisPost = posts[i];
    const author = thisPost.dataset['author'];

    let actions = thisPost.getElementsByClassName('article-actions').item(0);
    if (actions === null || !(actions instanceof HTMLElement)) {
      console.error('__');
    }

    let content = thisPost.getElementsByClassName('article-content').item(0);
    if (content === null || !(content instanceof HTMLElement)) {
      console.error('__');
    }

    let htmlContent = <HTMLElement>content;

    // Get the post's Markdown content, parsing it and replacing it with the rendered HTML.
    const trimmedContent = content.textContent.trim();

    //content.innerHTML = md.render(trimmedContent);
    
    let rendered = document.createElement('div');
    rendered.className = 'article-rendered';
    rendered.innerHTML = md.render(trimmedContent);
    
    htmlContent.style.display = 'none';

    thisPost.appendChild(rendered);

    let menuButton = document.createElement('button');
    menuButton.className = 'article-menu';
    menuButton.innerHTML = icon.menu;

    // Add 'Reply' and 'Option' buttons to each post, attaching handlers to them.
    let replyButton = document.createElement('button');
    replyButton.className = 'article-reply';
    replyButton.innerHTML = icon.reply;
    
    replyButton.addEventListener('click', handleReplyClick);

    let fragment = document.createDocumentFragment();
    fragment.appendChild(menuButton);
    fragment.appendChild(replyButton);

    if (isAdmin || (currentUser === author)) {
      let editButton = document.createElement('button');
      editButton.className = 'article-edit';
      editButton.innerHTML = icon.edit;
      editButton.addEventListener('click', handleEditClick);

      let deleteButton = document.createElement('button');
      deleteButton.className = 'article-delete';
      deleteButton.innerHTML = icon.del;
      deleteButton.addEventListener('click', handleDeleteClick);

      fragment.appendChild(deleteButton);
      fragment.appendChild(editButton);
    }

    actions.appendChild(fragment);
  }

  console.timeEnd('DOM_begin');
});