import React from "react";

interface FormMessageProps {
  type: "error" | "success" | "info" | "warning";
  message: string | null;
  className?: string;
}

/**
 * Component for displaying form messages (errors, success, etc.)
 */
export function FormMessage({
  type,
  message,
  className = "",
}: FormMessageProps) {
  if (!message) return null;

  const baseClasses = "px-4 py-3 rounded relative mb-4";

  const typeClasses = {
    error: "bg-red-100 border border-red-400 text-red-700",
    success: "bg-green-100 border border-green-400 text-green-700",
    info: "bg-blue-100 border border-blue-400 text-blue-700",
    warning: "bg-yellow-100 border border-yellow-400 text-yellow-700",
  };

  return (
    <div className={`${baseClasses} ${typeClasses[type]} ${className}`}>
      <span className="block sm:inline">{message}</span>
    </div>
  );
}

/**
 * Error message variant
 */
export function ErrorMessage({
  message,
  className,
}: Omit<FormMessageProps, "type">) {
  return <FormMessage type="error" message={message} className={className} />;
}

/**
 * Success message variant
 */
export function SuccessMessage({
  message,
  className,
}: Omit<FormMessageProps, "type">) {
  return <FormMessage type="success" message={message} className={className} />;
}

/**
 * Info message variant
 */
export function InfoMessage({
  message,
  className,
}: Omit<FormMessageProps, "type">) {
  return <FormMessage type="info" message={message} className={className} />;
}

/**
 * Warning message variant
 */
export function WarningMessage({
  message,
  className,
}: Omit<FormMessageProps, "type">) {
  return <FormMessage type="warning" message={message} className={className} />;
}
