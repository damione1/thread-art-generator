import htmx from 'htmx.org';
import Alpine from 'alpinejs';

window.htmx = htmx;
window.Alpine = Alpine;

function csrfToken() {
  return document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || '';
}

function bindHtmxCsrf() {
  document.body.addEventListener('htmx:configRequest', (evt) => {
    const csrf = csrfToken();
    if (csrf) {
      evt.detail.headers['X-CSRF-Token'] = csrf;
    }
  });
}

if (document.body) {
  bindHtmxCsrf();
} else {
  document.addEventListener('DOMContentLoaded', bindHtmxCsrf);
}

Alpine.start();
