type AuthResponse = {
  success?: boolean;
  message?: string;
};

function showError(message: string): void {
  const box = document.getElementById('auth-error');
  const text = document.getElementById('auth-error-message');
  if (!box || !text) {
    window.alert(message);
    return;
  }
  text.textContent = message;
  box.classList.remove('hidden');
}

async function postAuth(url: string, body: Record<string, string>): Promise<void> {
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(body),
  });
  const data = (await res.json().catch(() => ({}))) as AuthResponse;
  if (!res.ok || !data.success) {
    throw new Error(data.message || 'Authentication failed');
  }
}

function bindAuthForms(): void {
  document.getElementById('email-signin-form')?.addEventListener('submit', async (event) => {
    event.preventDefault();
    const form = event.target as HTMLFormElement;
    const email = (form.elements.namedItem('email') as HTMLInputElement)?.value ?? '';
    const password = (form.elements.namedItem('password') as HTMLInputElement)?.value ?? '';
    try {
      await postAuth('/auth/login', { email, password });
      window.location.href = '/dashboard';
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Sign in failed');
    }
  });

  document.getElementById('email-signup-form')?.addEventListener('submit', async (event) => {
    event.preventDefault();
    const form = event.target as HTMLFormElement;
    const email = (form.elements.namedItem('email') as HTMLInputElement)?.value ?? '';
    const password = (form.elements.namedItem('password') as HTMLInputElement)?.value ?? '';
    const confirm = (form.elements.namedItem('confirmPassword') as HTMLInputElement)?.value ?? '';
    if (password !== confirm) {
      showError('Passwords do not match');
      return;
    }
    try {
      await postAuth('/auth/signup', { email, password });
      window.location.href = '/dashboard';
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Sign up failed');
    }
  });
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', bindAuthForms);
} else {
  bindAuthForms();
}
