import { useState, useCallback, ChangeEvent, FormEvent } from "react";
import { handleFormError } from "../lib/grpc-error";

interface FormState<T> {
    values: T;
    errors: { [K in keyof T]?: string };
    touched: { [K in keyof T]?: boolean };
    isSubmitting: boolean;
    generalError: string | null;
    isSuccess: boolean;
}

type ValidationFunction<T> = (values: T) => { [K in keyof T]?: string };

export type UseFormResult<T> = {
    values: T;
    errors: { [K in keyof T]?: string };
    touched: { [K in keyof T]?: boolean };
    generalError: string | null;
    isSubmitting: boolean;
    isSuccess: boolean;
    handleChange: (
        e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
    ) => void;
    handleBlur: (
        e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
    ) => void;
    setFieldValue: (field: keyof T, value: unknown) => void;
    setFieldError: (field: keyof T, error: string) => void;
    setGeneralError: (error: string | null) => void;
    setFieldTouched: (field: keyof T, isTouched: boolean) => void;
    resetForm: () => void;
    validateForm: () => boolean;
    handleSubmit: (
        onSubmit: (values: T) => Promise<void>,
        options?: { resetOnSuccess?: boolean }
    ) => (e: FormEvent) => Promise<void>;
};

/**
 * Custom hook for form state management with error handling
 * @param initialValues - Initial form values
 * @param validate - Optional validation function
 * @returns Form state and handlers
 */
export function useForm<T extends Record<string, unknown>>(
    initialValues: T,
    validate?: ValidationFunction<T>
): UseFormResult<T> {
    const [state, setState] = useState<FormState<T>>({
        values: initialValues,
        errors: {},
        touched: {},
        isSubmitting: false,
        generalError: null,
        isSuccess: false,
    });

    const setFieldValue = useCallback(
        (field: keyof T, value: unknown) => {
            setState((prev) => ({
                ...prev,
                values: { ...prev.values, [field]: value },
                generalError: null,
                isSuccess: false,
                // Clear error for this field when value changes
                errors: {
                    ...prev.errors,
                    [field]: undefined,
                },
            }));
        },
        []
    );

    const setFieldError = useCallback(
        (field: keyof T, error: string) => {
            setState((prev) => ({
                ...prev,
                errors: { ...prev.errors, [field]: error },
            }));
        },
        []
    );

    const setGeneralError = useCallback(
        (error: string | null) => {
            setState((prev) => ({
                ...prev,
                generalError: error,
            }));
        },
        []
    );

    const setFieldTouched = useCallback(
        (field: keyof T, isTouched: boolean = true) => {
            setState((prev) => ({
                ...prev,
                touched: { ...prev.touched, [field]: isTouched },
            }));
        },
        []
    );

    const handleChange = useCallback(
        (
            e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
        ) => {
            const { name, value, type } = e.target as HTMLInputElement;
            const val = type === "checkbox" ? (e.target as HTMLInputElement).checked : value;
            setFieldValue(name as keyof T, val);
        },
        [setFieldValue]
    );

    const handleBlur = useCallback(
        (
            e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
        ) => {
            const { name } = e.target;
            setFieldTouched(name as keyof T, true);
        },
        [setFieldTouched]
    );

    const validateForm = useCallback((): boolean => {
        if (!validate) return true;

        const validationErrors = validate(state.values);
        const hasErrors = Object.keys(validationErrors).length > 0;

        setState((prev) => ({
            ...prev,
            errors: validationErrors,
        }));

        return !hasErrors;
    }, [state.values, validate]);

    const resetForm = useCallback(() => {
        setState({
            values: initialValues,
            errors: {},
            touched: {},
            isSubmitting: false,
            generalError: null,
            isSuccess: false,
        });
    }, [initialValues]);

    const handleSubmit = useCallback(
        (
            onSubmit: (values: T) => Promise<void>,
            options: { resetOnSuccess?: boolean } = { resetOnSuccess: false }
        ) => {
            return async (e: FormEvent): Promise<void> => {
                e.preventDefault();

                // Validate if validate function is provided
                const isValid = validateForm();
                if (!isValid) return;

                setState((prev) => ({
                    ...prev,
                    isSubmitting: true,
                    generalError: null,
                    isSuccess: false,
                }));

                try {
                    await onSubmit(state.values);

                    setState((prev) => ({
                        ...prev,
                        isSubmitting: false,
                        isSuccess: true,
                        ...(options.resetOnSuccess ? { values: initialValues, errors: {}, touched: {} } : {}),
                    }));
                } catch (err) {
                    console.error("Form submission error:", err);
                    // Handle gRPC errors with field violations
                    const handled = handleFormError(
                        err,
                        setFieldError,
                        setGeneralError
                    );

                    if (!handled) {
                        // Fallback error handling for non-gRPC errors
                        let errorMessage = "An unknown error occurred";
                        if (err instanceof Error) {
                            errorMessage = err.message;
                        }
                        setGeneralError(errorMessage);
                    }

                    setState((prev) => ({
                        ...prev,
                        isSubmitting: false,
                    }));
                }
            };
        },
        [state.values, validateForm, initialValues, setFieldError, setGeneralError]
    );

    return {
        values: state.values,
        errors: state.errors,
        touched: state.touched,
        generalError: state.generalError,
        isSubmitting: state.isSubmitting,
        isSuccess: state.isSuccess,
        handleChange,
        handleBlur,
        setFieldValue,
        setFieldError,
        setGeneralError,
        setFieldTouched,
        resetForm,
        validateForm,
        handleSubmit,
    };
}
