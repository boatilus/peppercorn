const md = window.markdownit();

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

const shiftKey   = 16;
const leftArrow  = 37;
const upArrow    = 38;
const rightArrow = 39;
const downArrow  = 40;

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

document.addEventListener('DOMContentLoaded', function() {
  let shifted = false;

  let bottom = document.getElementById('bottom');

  document.addEventListener('keydown', function(e) {
    const code = e.keyCode;

    if (bottom === document.activeElement) {
      return false;
    }

    if (code === shiftKey) {
      shifted = true;
      return;
    }

    if (code === leftArrow) {
      console.log('previous');
      return;
    }

    if (code === rightArrow) {
      console.log('next');
      return;
    }

    if (code === upArrow) {
      if (shifted) window.scrollTo(0, 0);
      return;
    }

    if (code === downArrow) {
      if (shifted) {
        window.scrollTo(0, document.body.clientHeight);
        bottom.focus();
      }
    }
  });

  document.addEventListener('keyup', function(e) {
    if (e.keyCode === shiftKey) shifted = false;
  });

  console.time('DOM_begin');
  
  let posts = document.getElementsByTagName('article');

  for (let i = 0; i < posts.length; i++) {
    let thisPost = posts[i];

    let actions = thisPost.getElementsByClassName('article-actions').item(0);
    let content = thisPost.getElementsByClassName('article-content').item(0);

    // Get the post's Markdown content, parsing it and replacing it with the rendered HTML.
    const trimmedContent = content.textContent.trim();

    content.innerHTML = md.render(trimmedContent);

    // Add 'Reply' and 'Option' buttons to each post, attaching handlers to them.
    let replyButton = document.createElement('button');
    replyButton.className = 'article-reply';
    replyButton.innerHTML = replyIcon;
    
    replyButton.addEventListener('click', function() {
      const strippedAndQuoted = quote(stripQuotes(trimmedContent));

      bottom.value = strippedAndQuoted;
      bottom.focus();
    });

    let menuButton = document.createElement('button');
    menuButton.className = 'article-menu';
    menuButton.innerHTML = menuIcon;

    let editButton = document.createElement('button');
    editButton.className = 'article-edit';
    editButton.innerHTML = editIcon;

    let deleteButton = document.createElement('button');
    deleteButton.className = 'article-delete';
    deleteButton.innerHTML = deleteIcon;
    deleteButton.addEventListener('click', function() {
      const ok = confirm('Are you sure you want to delete this post?');

      if (ok) {
        // Change URL to delete request
      }
    });

    let fragment = document.createDocumentFragment();
    fragment.appendChild(deleteButton);
    fragment.appendChild(editButton);
    fragment.appendChild(menuButton);
    fragment.appendChild(replyButton);

    actions.appendChild(fragment);
  }

  console.timeEnd('DOM_begin');
});