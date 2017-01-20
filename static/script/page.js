const md = window.markdownit();

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
  console.time('DOM_begin');

  let bottom = document.getElementById('bottom');
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
    replyButton.innerHTML =
      `<svg version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" preserveAspectRatio="xMidYMid meet" viewBox="0 0 384 320">
        <path class="fill" d="M149.333,85.333 C298.667,106.667 362.667,213.333 384,320 C330.667,245.333 256,211.2 149.333,211.2 L149.333,298.667 L0,149.333 L149.333,0 L149.333,85.333 z" />
      </svg>`;
    
    replyButton.addEventListener('click', function() {
      const strippedAndQuoted = quote(stripQuotes(trimmedContent));

      bottom.value = strippedAndQuoted;
      bottom.focus();
    });

    let menuButton = document.createElement('button');
    menuButton.className = 'article-menu';
    menuButton.innerHTML =
      `<svg version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" preserveAspectRatio="xMidYMid meet" viewBox="0 0 224 56">
        <g class="fill">
          <path d="M28,56 C12.536,56 -0,43.464 -0,28 C-0,12.536 12.536,0 28,0 C43.464,0 56,12.536 56,28 C56,43.464 43.464,56 28,56 z" />
          <path d="M112,56 C96.536,56 84,43.464 84,28 C84,12.536 96.536,0 112,0 C127.464,0 140,12.536 140,28 C140,43.464 127.464,56 112,56 z" />
          <path d="M196,56 C180.536,56 168,43.464 168,28 C168,12.536 180.536,0 196,0 C211.464,0 224,12.536 224,28 C224,43.464 211.464,56 196,56 z" />
        </g>
      </svg>`;

    let fragment = document.createDocumentFragment();
    fragment.appendChild(menuButton);
    fragment.appendChild(replyButton);

    actions.appendChild(fragment);
  }

  console.timeEnd('DOM_begin');
});