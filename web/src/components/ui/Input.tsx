import React, { InputHTMLAttributes } from "react";

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  id: string;
  name: string;
  error?: string;
  fullWidth?: boolean;
}

/**
 * A flexible Input component with consistent styling
 * Can be used standalone or with FormField
 */
export function Input({
  id,
  name,
  error,
  className = "",
  fullWidth = true,
  ...props
}: InputProps) {
  const baseClasses =
    "p-2 bg-dark-300 border rounded text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500";
  const errorClasses = error ? "border-red-500" : "border-dark-400";
  const widthClasses = fullWidth ? "w-full" : "";

  return (
    <input
      id={id}
      name={name}
      className={`${baseClasses} ${errorClasses} ${widthClasses} ${className}`}
      {...props}
    />
  );
}

/**
 * Text input variant
 */
export function TextInput(props: InputProps) {
  return <Input type="text" {...props} />;
}

/**
 * Email input variant
 */
export function EmailInput(props: InputProps) {
  return <Input type="email" {...props} />;
}

/**
 * Password input variant
 */
export function PasswordInput(props: InputProps) {
  return <Input type="password" {...props} />;
}

/**
 * URL input variant
 */
export function UrlInput(props: InputProps) {
  return <Input type="url" {...props} />;
}
