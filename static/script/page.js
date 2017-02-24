const spoilerReg = new RegExp(/^spoiler\s+(.*)$/);

const md = new markdownit({
  html: true,
  linkify: true,    // Automatically convert URLs to links.
  typographer: true // Automatically convert some tokens, like (C) to the copyright symbol.
}).use(markdownitContainer, 'spoiler', {
  validate: function(params) {
    return params.trim().match(spoilerReg);
  },

  render: function(tokens, idx) {
    var m = tokens[idx].info.trim().match(spoilerReg);

    if (tokens[idx].nesting === 1) {
      // opening tag
      return '<button class="article-spoiler-button">Spoiler: ' + md.utils.escapeHtml(m[1]) + '</button><div class="article-spoiler">';
    } else {
      // closing tag
      return '</div>';
    }
  }
});

const menuIcon = 
  `<svg
    version="1.1"
    xmlns="http://www.w3.org/2000/svg"
    xmlns:xlink="http://www.w3.org/1999/xlink"
    preserveAspectRatio="xMidYMid meet"
    viewBox="0 0 224 56"
  >
    <g class="fill">
      <path d="M28,56 C12.536,56 -0,43.464 -0,28 C-0,12.536 12.536,0 28,0 C43.464,0 56,12.536 56,28 C56,43.464 43.464,56 28,56 z" />
      <path d="M112,56 C96.536,56 84,43.464 84,28 C84,12.536 96.536,0 112,0 C127.464,0 140,12.536 140,28 C140,43.464 127.464,56 112,56 z" />
      <path d="M196,56 C180.536,56 168,43.464 168,28 C168,12.536 180.536,0 196,0 C211.464,0 224,12.536 224,28 C224,43.464 211.464,56 196,56 z" />
    </g>
  </svg>`;

const replyIcon =
  `<svg
    version="1.1"
    xmlns="http://www.w3.org/2000/svg"
    xmlns:xlink="http://www.w3.org/1999/xlink"
    preserveAspectRatio="xMidYMid meet"
    viewBox="0 0 384 320"
  >
    <path class="fill" d="M149.333,85.333 C298.667,106.667 362.667,213.333 384,320 C330.667,245.333 256,211.2 149.333,211.2 L149.333,298.667 L0,149.333 L149.333,0 L149.333,85.333 z" />
  </svg>`;

const editIcon =
  `<svg
    version="1.1" 
    xmlns="http://www.w3.org/2000/svg"
    xmlns:xlink="http://www.w3.org/1999/xlink"
    preserveAspectRatio="xMidYMid meet"
    viewBox="0 0 528.899 528.899"
  >
    <path class="fill" d="M328.883,89.125l107.59,107.589l-272.34,272.34L56.604,361.465L328.883,89.125z M518.113,63.177l-47.981-47.981
      c-18.543-18.543-48.653-18.543-67.259,0l-45.961,45.961l107.59,107.59l53.611-53.611
      C532.495,100.753,532.495,77.559,518.113,63.177z M0.3,512.69c-1.958,8.812,5.998,16.708,14.811,14.565l119.891-29.069
      L27.473,390.597L0.3,512.69z" />
  </svg>`

const deleteIcon =
  `<svg
    version="1.1"
    xmlns="http://www.w3.org/2000/svg"
    xmlns:xlink="http://www.w3.org/1999/xlink"
    viewBox="0, 0, 360, 360"
    >
      <circle class="red" cx="180" cy="180" r="180" />
      <path
        class="white"
        d="M243.585,88 L272,116.415 L208.415,180 L272,243.585 L243.585,272 L180,208.415
          L116.415,272 L88,243.585 L151.586,180 L88,116.415 L116.415,88 L180,151.586 L243.585,88 z"
        />
    </g>
  </svg>`

const returnKey  = 13;
const shiftKey   = 16;
const escapeKey  = 27;
const leftArrow  = 37;
const upArrow    = 38;
const rightArrow = 39;
const downArrow  = 40;
const hKey       = 72;

let isAdmin     = false;
let currentUser = '';

// The following are selectors for which we need to cache, as we'll reference their existence in
// key handlers.
let prev   = null;
let next   = null;
let reply  = null;
let bottom = null;
let modal  = null;

// Retrieve the nearest ancestor that matches `tag`, returning `null` if it hits the <html> element.
Element.prototype.getAncestorByTagName = function(tag) {
  let e = this;

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
}

KeyboardEvent.prototype.isModified = function() {
  return this.ctrlKey || this.metaKey;
}

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

// Binds `onclick` handlers for spoiler buttons within HTMLElement `node`
const bindSpoilersFor = function(node) {
  if (!node instanceof HTMLElement || node === null) {
    console.error('bindSpoilersFor: argument is not of type HTMLElement or null');
    return;
  }

  let spoilerButtons = node.getElementsByClassName('article-spoiler-button');
  if (spoilerButtons.length === 0) return;

  for (let i = 0; i < spoilerButtons.length; i++) {
    spoilerButtons[i].addEventListener('click', handleSpoilerClick);
  }
}

const handleKeyDownEvents = function(event) {
  const nodeName = document.activeElement.nodeName;
  const isInputFocused = nodeName === 'TEXTAREA' || nodeName === 'INPUT';

  // We want users to be able to hit Esc and get out of the reply field so they can get back to
  // using shortcuts for everything else.
  if (isInputFocused) {
    if (event.keyCode === escapeKey) {
      document.activeElement.blur();
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
    if (event.shiftKey && (next !== null)) window.location = next.getAttribute('href');
    return;
  case upArrow:
    if (event.shiftKey) window.scrollTo(0, 0);
    return;
  case downArrow:
    if (event.shiftKey) {
      window.scrollTo(0, document.body.clientHeight);
      bottom.focus();
    }
    return;
  case hKey:
    if (event.shiftKey) {
      let spoilers = document.getElementsByClassName('article-spoiler-button');
      if (spoilers.length === 0) return;

      for (let i = 0; i < spoilers.length; i++) {
        spoilers[i].style.display = 'block';
        spoilers[i].nextSibling.style.display = 'none';
      }
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

const handleMenuClick = function() {
  const id = this.dataset['id'];

  if (typeof id === 'undefined') {
    console.error('handleMenuClick: no "id" data attribute bound on edit button');
    return;
  }
  
  if (modal === null) {
    console.error('handleMenuClick: could not find modal element');
    return;
  }

  // Clear the contents first, before proceding with handler binding.
  for (; modal.children.length; ) {
    modal.children.item(0).remove();
  }

  let fragment = document.createDocumentFragment();

  let edit = document.createElement('li');
  edit.className     = 'article-menu-modal-edit';
  edit.dataset['id'] = id;
  edit.innerText     = 'Edit';
  edit.addEventListener('click', handleEditClick);

  let del = document.createElement('li');
  del.className = 'delete';
  del.dataset['id'] = id;
  del.innerText     = 'Delete';
  del.addEventListener('click', handleDeleteClick);

  fragment.appendChild(edit);
  fragment.appendChild(del);
  modal.appendChild(fragment);

  const bodyRect       = document.body.getBoundingClientRect();
  const menuButtonRect = this.getBoundingClientRect();

  const offsetY = menuButtonRect.top  - bodyRect.top;
  const offsetX = menuButtonRect.left - bodyRect.left;

  modal.style.top  = offsetY + this.offsetHeight + 'px';
  modal.style.left = offsetX - this.offsetWidth  + 'px';
  modal.style.display = 'inline-block';
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

    submitButton.remove();
    submitButton = null;

    action.style.visibility = 'visible';
    rendered.style.display  = 'block';
    modal.style.display     = 'none';
  };
  
  // We'll create a textarea element filled with the post's Markdown comment right within the
  // post, then temporarily hide the rendered content. On Ctrl+Enter or ⌘+Enter, we'll PATCH
  // the post with the new content and re-render with revised content on success. If the user hits
  // Esc, we'll simply remove the textarea element and show the existing rendered content.
  const id = this.dataset['id'];

  if (typeof id === 'undefined') {
    console.error('handleEditClick: no "id" data attribute bound on edit button');
    return false;
  }

  const article = document.getElementById(id);
  if (article === null) {
    console.error('handleEditClick: could not find article with id ' + id);
    return false;
  }

  const actions = article.getElementsByClassName('article-actions');
  if (actions.length === 0) {
    console.error('handleEditClick: no actions group found for this post');
    return false;
  }

  let action = actions.item(0);

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

  // fragment will hold the textarea and the submit button.
  let fragment = document.createDocumentFragment();
  
  let editable = document.createElement('textarea');
  editable.className = 'article-editable';
  editable.value     = content.textContent;
  editable.rows      = 6;

  // Create a submit button to display on mobile, as we're not creating a full for this.
  let submitButton = document.createElement('button');
  submitButton.id        = 'article-edit-submit';
  submitButton.innerText = 'Submit changes';

  const submitEdit = function() {
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

      content.innerHTML  = val;

      // Keep the new Markdown in `article-content`, and render it to `article-rendered`.
      rendered.remove();
      rendered = null; // To mark for collection.
      
      rendered = document.createElement('div');
      rendered.className = 'article-rendered';
      rendered.innerHTML = md.render(val);

      bindSpoilersFor(rendered);
      article.appendChild(rendered);

      displayViewState();
    });
    xhr.addEventListener('timeout', function() {
      console.timeEnd('post-edit');
      console.error("handleEditClick: PATCH request timed out");
    });
    xhr.send(JSON.stringify({ content: editable.value }));
  }

  const handleKeydown = function(e) {
    if (e.keyCode === escapeKey) {
      displayViewState();
      return;
    }

    if (e.isModified() && (e.keyCode === returnKey)) {
      submitEdit();
    }
  }

  editable.addEventListener('keydown', handleKeydown);
  submitButton.addEventListener('click', submitEdit);

  fragment.appendChild(editable);
  fragment.appendChild(submitButton);
  article.appendChild(fragment);
  editable.focus();

  action.style.visibility = 'hidden';
  modal.style.display      = 'none';
  rendered.style.display   = 'none';
};

const handleDeleteClick = function(event) {
  const id = this.dataset['id'];

  if (typeof id === 'undefined') {
    console.error('handleDeleteClick: no "id" data attribute bound on edit button');
    return false;
  }

  const article = document.getElementById(id);
  if (article === null) {
    console.error('handleDeleteClick: could not find article ancestor with id ' + id);
    return false;
  }

  if (window.confirm('Are you sure you want to delete this post?')) {
    window.location.href = `/posts/${article.id}/delete`;
  }
};

const handleSpoilerClick = function() {
  this.style.display = 'none';
  this.nextSibling.style.display = 'block';
}

document.addEventListener('DOMContentLoaded', function() {
  isAdmin     = (document.body.dataset['isAdmin'] === 'true');
  currentUser = document.body.dataset['currentUser'];

  prev   = document.getElementById('nav-previous');
  next   = document.getElementById('nav-next');
  reply  = document.getElementById('reply');
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

  // Add a listener to submit a reply on Ctrl+Enter/Option+Enter
  bottom.addEventListener('keydown', function(e) {
    if (bottom.value != "" && e.isModified() && (e.keyCode === returnKey)) {
      reply.submit();
    }
  })

  // Add a div to contain the post menu, which we'll show/hide and move around as necessary.
  modal = document.createElement('ul');
  modal.id = 'article-menu-modal';

  let bodyFragment = document.createDocumentFragment();
  bodyFragment.appendChild(prevArrow);
  bodyFragment.appendChild(nextArrow);
  bodyFragment.appendChild(modal);

  document.body.appendChild(bodyFragment);

  window.addEventListener('keydown', handleKeyDownEvents);
  window.addEventListener('keyup', handleKeyUpEvents);

  console.time('DOM_begin');
  
  let posts = document.getElementsByTagName('article');

  for (let i = 0; i < posts.length; i++) {
    const thisPost = posts[i];
    const author = thisPost.dataset.author;

    let actions = thisPost.getElementsByClassName('article-actions').item(0);
    let content = thisPost.getElementsByClassName('article-content').item(0);

    // Get the post's Markdown content, parsing it and replacing it with the rendered HTML.
    const trimmedContent = content.textContent;
    
    let rendered = document.createElement('div');
    rendered.className = 'article-rendered';
    rendered.innerHTML = md.render(trimmedContent);
    
    content.style.display = 'none';

    thisPost.appendChild(rendered);

    let menuButton = document.createElement('button');
    menuButton.className = 'article-menu';
    menuButton.innerHTML = menuIcon;
    menuButton.dataset['id'] = thisPost.id;
    menuButton.addEventListener('click', handleMenuClick);

    // Add 'Reply' and 'Menu' buttons to each post, attaching handlers to them.
    let replyButton = document.createElement('button');
    replyButton.className = 'article-reply';
    replyButton.innerHTML = replyIcon;
    replyButton.addEventListener('click', handleReplyClick);

    let fragment = document.createDocumentFragment();
    if (isAdmin || (currentUser === author)) {
      fragment.appendChild(menuButton);
    }
    
    fragment.appendChild(replyButton);

    if (isAdmin || (currentUser === author)) {
      let editButton = document.createElement('button');
      editButton.className = 'article-edit';
      editButton.innerHTML = editIcon;
      editButton.dataset['id'] = thisPost.id;
      editButton.addEventListener('click', handleEditClick);

      let deleteButton = document.createElement('button');
      deleteButton.className = 'article-delete';
      deleteButton.innerHTML = deleteIcon;
      deleteButton.dataset['id'] = thisPost.id;
      deleteButton.addEventListener('click', handleDeleteClick);

      fragment.appendChild(deleteButton);
      fragment.appendChild(editButton);
    }

    actions.appendChild(fragment);

    bindSpoilersFor(thisPost);
  }

  console.timeEnd('DOM_begin');
});