# Form Components Library

A flexible, reusable form components library for React applications using TypeScript and Tailwind CSS.

## Components

### FormField

A wrapper component that provides consistent layout and styling for form inputs with labels and error messages.

```tsx
<FormField
  id="firstName"
  name="firstName"
  label="First Name"
  error={errors.firstName}
  required={true}
  hint="Enter your legal first name"
  inputComponent={
    <TextInput
      id="firstName"
      name="firstName"
      value={values.firstName}
      onChange={handleChange}
      disabled={isSubmitting}
    />
  }
/>
```

### Input Components

- `Input`: Base input component with consistent styling
- `TextInput`: Text input variant
- `EmailInput`: Email input variant
- `PasswordInput`: Password input variant
- `UrlInput`: URL input variant

```tsx
<TextInput
  id="firstName"
  name="firstName"
  value={values.firstName}
  onChange={handleChange}
  disabled={isSubmitting}
/>
```

### Message Components

- `FormMessage`: Base message component
- `ErrorMessage`: Error message variant
- `SuccessMessage`: Success message variant
- `InfoMessage`: Info message variant
- `WarningMessage`: Warning message variant

```tsx
<ErrorMessage message={generalError} />
<SuccessMessage message={isSuccess ? "Form submitted successfully!" : null} />
```

## useForm Hook

A powerful hook for managing form state, validation, and submission.

```tsx
const {
  values,
  errors,
  touched,
  handleChange,
  handleBlur,
  handleSubmit,
  setFieldValue,
  setFieldError,
  setGeneralError,
  isSubmitting,
  generalError,
  isSuccess,
  resetForm,
} = useForm<YourFormValuesType>({
  firstName: "",
  lastName: "",
  email: "",
});
```

## Example Usage

```tsx
import {
  FormField,
  TextInput,
  EmailInput,
  ErrorMessage,
  SuccessMessage,
  useForm,
} from "@/components/ui";

interface MyFormValues extends Record<string, unknown> {
  firstName: string;
  lastName: string;
  email: string;
}

export default function MyForm() {
  const {
    values,
    errors,
    handleChange,
    handleSubmit,
    isSubmitting,
    generalError,
    isSuccess,
  } = useForm<MyFormValues>({
    firstName: "",
    lastName: "",
    email: "",
  });

  const onSubmitForm = async (formValues: MyFormValues) => {
    // Submit your form data
    await submitData(formValues);
  };

  return (
    <div className="p-6">
      <h2 className="text-xl font-semibold mb-4">My Form</h2>

      <ErrorMessage message={generalError} />
      <SuccessMessage
        message={isSuccess ? "Form submitted successfully!" : null}
      />

      <form onSubmit={handleSubmit(onSubmitForm)}>
        <div className="grid grid-cols-2 gap-4 mb-4">
          <FormField
            id="firstName"
            name="firstName"
            label="First Name"
            error={errors.firstName}
            inputComponent={
              <TextInput
                id="firstName"
                name="firstName"
                value={values.firstName}
                onChange={handleChange}
                disabled={isSubmitting}
              />
            }
          />

          <FormField
            id="lastName"
            name="lastName"
            label="Last Name"
            error={errors.lastName}
            inputComponent={
              <TextInput
                id="lastName"
                name="lastName"
                value={values.lastName}
                onChange={handleChange}
                disabled={isSubmitting}
              />
            }
          />

          <FormField
            id="email"
            name="email"
            label="Email"
            error={errors.email}
            inputComponent={
              <EmailInput
                id="email"
                name="email"
                value={values.email}
                onChange={handleChange}
                disabled={isSubmitting}
              />
            }
          />
        </div>

        <button
          type="submit"
          disabled={isSubmitting}
          className="px-4 py-2 bg-blue-600 text-white rounded"
        >
          {isSubmitting ? "Submitting..." : "Submit"}
        </button>
      </form>
    </div>
  );
}
```
