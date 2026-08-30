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

function solverContrastFromInput(value) {
  return String((100 + Number(value)) / 100);
}

function bindSolverContrastPreview() {
  document.body.addEventListener('input', (evt) => {
    const target = evt.target;
    if (!(target instanceof HTMLInputElement) || target.id !== 'image_contrast') {
      return;
    }
    const editor = target.closest('#composition-editor');
    const disc = editor?.querySelector('.disc-photo');
    if (disc instanceof HTMLElement) {
      disc.style.setProperty('--solver-contrast', solverContrastFromInput(target.value));
    }
    editor?.querySelectorAll('[data-contrast-readout]').forEach((el) => {
      el.textContent = `${target.value}%`;
    });
  });

  document.body.addEventListener('htmx:afterSwap', (evt) => {
    const el = evt.detail?.elt;
    if (el instanceof Element) {
      Alpine.initTree(el);
    }
  });
}

if (document.body) {
  bindHtmxCsrf();
  bindSolverContrastPreview();
} else {
  document.addEventListener('DOMContentLoaded', () => {
    bindHtmxCsrf();
    bindSolverContrastPreview();
  });
}

Alpine.start();
