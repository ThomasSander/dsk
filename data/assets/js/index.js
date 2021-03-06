/*!
 * Copyright 2017 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

document.addEventListener('DOMContentLoaded', function() {
  let $ = document.querySelectorAll.bind(document);
  let $1 = document.querySelector.bind(document);

  let nav = $1('.tree-nav');
  let search = $1('.search-field');

  let tree = new Tree();

  let fuse = new Fuse([], {
    tokenize: false,
    matchAllTokens: true,
    threshold: 0.1,
    location: 0,
    distance: 100,
    maxPatternLength: 32,
    minMatchCharLength: 1,
    keys: [
      "title",
      "url",
      "meta.keywords"
    ]
  });

  // Gets the tree and creates the nav structure.
  // Get the query from the current window path (handleSearchWithQuery will render the Nav).
  tree.sync()
    .then(() => {
      fuse.setCollection(tree.flatten());
      handleSearchWithQuery(window.location.search.substring(1));
    });

  // Loads the node based on url
  let handleUrl = function(url) {
    fetch(url).then((res) => {
      return res.text();
    }).then((html) => {
      $1('main').innerHTML = html;
      handleKeywords();
    });
  };

  // Initial check for route and load node
  if (window.location.pathname !== '/') {
    let url = window.location.protocol + '//' +
      window.location.host + '/tree' +
      window.location.pathname;
    handleUrl(url);
  }

  // Runs the search from the input field
  let handleSearch = function(ev) {
    // Add query to the url
    let uri = window.location.origin + window.location.pathname + "?" + this.value;
    history.replaceState(null, '', uri);

    if (this.value !== "") {
      renderNav(tree, fuse.search(this.value));
    } else {
      renderNav(tree);
    }
  };

  // Runs the search with a given query
  let handleSearchWithQuery = function(q) {
    // Add query to the url
    let uri = window.location.origin + window.location.pathname + "?" + q;
    history.replaceState(null, '', uri);

    search.value = q;

    if (q !== "") {
      renderNav(tree, fuse.search(q));
    } else {
      renderNav(tree);
    }
  };

  // Clears the search field
  let clearSearch = function() {
    handleSearchWithQuery("");
  };

  search.addEventListener("input", handleSearch);
  $1('.search-clear').addEventListener("click", clearSearch);

  // Loads the node when a link the nav is clicked
  // and updates session history (uri)
  let handleNav = function(ev) {
    ev.preventDefault();
    fetch(this.href).then((res) => {
      return res.text();
    }).then((html) => {
      let uri = this.href.split('tree').pop() + window.location.search;
      history.pushState(null, '', uri);
      $1('main').innerHTML = html;
      handleKeywords();
    });
  };

  // Calls the search when a keyword is clicked
  let handleKeywordClick = function(ev) {
    handleSearchWithQuery(ev.target.innerHTML);
  };

  // Attaches a click-Event to every keyword
  let handleKeywords = function() {
    for (let k of $('.keyword')) {
      k.addEventListener("click", handleKeywordClick);
    }
  };

  // Renders the nav structure
  let renderNav = function(tree, searchResults) {
    nav.innerHTML = '';
    var list;

    if (searchResults) {
      list = createList(tree.filteredBy(searchResults).root);
    } else {
      // When none selected, all nodes should be kept, resets view.
      list = createList(tree.root);
    }
    if (list) {
      // Append list withouth root node.
      nav.appendChild(list.childNodes[1]);
    }
  };


  // Turns the given data into a "ul li" structure
  let createList = function(node) {
    if (node.keep === false) {
      return;
    }

    let li = document.createElement('li');
    let a  = document.createElement('a');
    a.href = '/tree/' + node.url;
    a.innerHTML = node.title;
    a.addEventListener('click', handleNav);

    if (node.isGhost) {
      a.classList.add('ghosted');
    }

    li.appendChild(a);

    let ul = document.createElement('ul');
    li.appendChild(ul);

    for (var child in node.children) {
      var childList = createList(node.children[child]);
      if (childList) {
        ul.appendChild(childList);
      }
    }

    return li;
  };
});
