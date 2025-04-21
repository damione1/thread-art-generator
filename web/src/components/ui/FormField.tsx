import React, { ReactNode } from "react";

interface FormFieldProps {
  id: string;
  name: string;
  label: string;
  error?: string;
  disabled?: boolean;
  required?: boolean;
  className?: string;
  labelClassName?: string;
  inputContainerClassName?: string;
  errorClassName?: string;
  inputComponent: ReactNode;
  hint?: string;
}

/**
 * A flexible FormField component that can wrap any input component
 * with consistent label, error, and hint display
 */
export function FormField({
  id,
  label,
  error,
  required = false,
  className = "",
  labelClassName = "block text-sm font-medium text-slate-400 mb-1",
  inputContainerClassName = "",
  errorClassName = "text-red-500 text-xs mt-1",
  inputComponent,
  hint,
}: FormFieldProps) {
  return (
    <div className={className}>
      <label htmlFor={id} className={labelClassName}>
        {label}
        {required && <span className="text-red-500 ml-1">*</span>}
      </label>

      <div className={inputContainerClassName}>{inputComponent}</div>

      {error && <p className={errorClassName}>{error}</p>}

      {hint && !error && <p className="text-slate-500 text-xs mt-1">{hint}</p>}
    </div>
  );
}
