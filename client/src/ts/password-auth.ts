type AuthResponse = {
  success?: boolean;
  message?: string;
};

function showError(message: string): void {
  const box = document.getElementById('auth-error');
  const text = document.getElementById('auth-error-message');
  document.getElementById('auth-success')?.classList.add('hidden');
  if (!box || !text) {
    window.alert(message);
    return;
  }
  text.textContent = message;
  box.classList.remove('hidden');
}

function showSuccess(message: string): void {
  const box = document.getElementById('auth-success');
  const text = document.getElementById('auth-success-message');
  document.getElementById('auth-error')?.classList.add('hidden');
  if (!box || !text) {
    return;
  }
  text.textContent = message;
  box.classList.remove('hidden');
}

async function postAuth(url: string, body: Record<string, string>): Promise<AuthResponse> {
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
  return data;
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
    const firstName = (form.elements.namedItem('first_name') as HTMLInputElement)?.value ?? '';
    const lastName = (form.elements.namedItem('last_name') as HTMLInputElement)?.value ?? '';
    if (password !== confirm) {
      showError('Passwords do not match');
      return;
    }
    try {
      await postAuth('/auth/signup', {
        email,
        password,
        first_name: firstName,
        last_name: lastName,
      });
      window.location.href = '/check-email';
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Sign up failed');
    }
  });

  document.getElementById('forgot-password-form')?.addEventListener('submit', async (event) => {
    event.preventDefault();
    const form = event.target as HTMLFormElement;
    const email = (form.elements.namedItem('email') as HTMLInputElement)?.value ?? '';
    try {
      const data = await postAuth('/auth/forgot-password', { email });
      showSuccess(data.message || 'If that email exists, we sent a reset link');
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Request failed');
    }
  });

  document.getElementById('resend-verification-form')?.addEventListener('submit', async (event) => {
    event.preventDefault();
    const form = event.target as HTMLFormElement;
    const email = (form.elements.namedItem('email') as HTMLInputElement)?.value ?? '';
    try {
      const data = await postAuth('/auth/resend-verification', { email });
      showSuccess(data.message || 'If that email exists, we sent a confirmation link');
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Request failed');
    }
  });
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', bindAuthForms);
} else {
  bindAuthForms();
}
