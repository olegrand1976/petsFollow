(function () {
  try {
    var m = document.cookie.match(/(?:^|;\s*)pf_theme=([^;]+)/);
    if (m && decodeURIComponent(m[1]) === 'dark') {
      document.documentElement.classList.add('dark');
    }
  } catch (e) {}
})();
