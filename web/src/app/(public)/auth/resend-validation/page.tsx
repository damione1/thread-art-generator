'use client';

import React, { useState, ChangeEvent, FormEvent, useEffect } from "react";
import Link from "next/link";
import Breadcrumb from "@/components/Breadcrumbs/Breadcrumb";
import { SendValidationEmailRequest } from "@/../grpc/user_pb";
import { ArtGeneratorServiceClient } from "@/../grpc/ServicesServiceClientPb";
import Toast from "@/components/Notification/Toast";
import { parseValidationErrors } from "@/utils/errorUtils";
import { useRouter } from "next/navigation";
import { saveEmail, getSavedEmail, clearSavedEmail } from "@/utils/authUtils";

const ResendValidation: React.FC = () => {
  const [errors, setErrors] = useState<{ [key: string]: string }>({});
  const [toast, setToast] = useState<{ message: string; type: 'error' | 'success' | 'info' } | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [email, setEmail] = useState('');
  const [debugInfo, setDebugInfo] = useState<{
    requestSent: boolean;
    requestTime: string;
    email: string;
    responseStatus: string;
    errorDetails: string;
  } | null>(null);
  const router = useRouter();

  // Load saved email on component mount
  useEffect(() => {
    const savedEmail = getSavedEmail();
    if (savedEmail) {
      setEmail(savedEmail);
    }
  }, []);

  // Handle input change to clear error for that field
  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;

    if (name === 'email') {
      setEmail(value);
    }

    if (errors[name]) {
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[name];
        return newErrors;
      });
    }
  };

  // Handle "Not you?" click
  const handleNotYou = () => {
    clearSavedEmail();
    setEmail('');
  };

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setErrors({});
    setToast(null);
    setIsSubmitting(true);
    setDebugInfo(null);

    // Save email for future use
    saveEmail(email);

    // Initialize debug info
    const debugData = {
      requestSent: false,
      requestTime: new Date().toISOString(),
      email: email,
      responseStatus: 'pending',
      errorDetails: ''
    };

    const client = new ArtGeneratorServiceClient(process.env.NEXT_PUBLIC_GRPC_API!);
    const request = new SendValidationEmailRequest();
    request.setEmail(email);

    try {
      console.log(`[DEBUG] Sending validation email to: ${email}`);
      console.log(`[DEBUG] API endpoint: ${process.env.NEXT_PUBLIC_GRPC_API}`);

      debugData.requestSent = true;

      const response = await client.sendValidationEmail(request);

      console.log(`[DEBUG] Email sent successfully, response:`, response);
      debugData.responseStatus = 'success';

      setToast({
        type: 'success',
        message: 'Validation email has been sent. Please check your inbox and spam folder.'
      });

      // Redirect to validate-email page after 2 seconds
      setTimeout(() => {
        router.push(`/auth/validate-email?email=${encodeURIComponent(email)}`);
      }, 2000);
    } catch (error: any) {
      console.error("[DEBUG] Error sending validation email:", error);
      debugData.responseStatus = 'error';

      // Extract more detailed error information
      const errorMessage = error.message as string;
      debugData.errorDetails = errorMessage;

      // Log additional error details if available
      if (error.code) console.error(`[DEBUG] Error code: ${error.code}`);
      if (error.metadata) console.error(`[DEBUG] Error metadata:`, error.metadata);

      // Parse validation errors using the utility function
      const validationErrors = parseValidationErrors(errorMessage);

      if (Object.keys(validationErrors).length > 0) {
        // Check for generic error
        if (validationErrors._generic) {
          setToast({
            type: 'error',
            message: validationErrors._generic
          });
          delete validationErrors._generic;
        }

        // Set field-specific errors
        if (Object.keys(validationErrors).length > 0) {
          setErrors(validationErrors);
        }
      } else {
        // Handle non-field specific errors
        setToast({
          type: 'error',
          message: 'An error occurred while sending the validation email. Please check your email address and try again later.'
        });
      }
    } finally {
      setIsSubmitting(false);
      setDebugInfo(debugData);
    }
  }

  // Determine if we're in development mode
  const isDevelopment = process.env.NODE_ENV === 'development';

  return (
    <>
      <Breadcrumb pageName="Resend Validation Email" />

      <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
        <div className="flex flex-wrap items-center">
          <div className="w-full border-stroke dark:border-strokedark xl:w-1/2 xl:border-l-2">
            <div className="w-full p-4 sm:p-12.5 xl:p-17.5">
              <h2 className="mb-9 text-2xl font-bold text-black dark:text-white sm:text-title-xl2">
                Resend Validation Email
              </h2>
              <p className="mb-6 text-base text-body-color dark:text-body-color-dark">
                Enter your email address below to receive a new validation email.
              </p>

              {toast && (
                <Toast
                  message={toast.message}
                  type={toast.type}
                  onClose={() => setToast(null)}
                />
              )}

              <form onSubmit={onSubmit}>
                <div className="mb-4">
                  <label className="mb-2.5 block font-medium text-black dark:text-white">
                    Email
                  </label>
                  <div className="relative">
                    <input
                      type="email"
                      name="email"
                      value={email}
                      placeholder="Enter your email"
                      className={`w-full rounded-lg border ${errors.email ? 'border-red-500' : 'border-stroke'} bg-transparent py-4 pl-6 pr-10 outline-none focus:border-primary focus-visible:shadow-none dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary`}
                      onChange={handleInputChange}
                      required
                    />
                    {errors.email && (
                      <span className="text-red-500 text-sm mt-1">{errors.email}</span>
                    )}
                    <span className="absolute right-4 top-4">
                      <svg
                        className="fill-current"
                        width="22"
                        height="22"
                        viewBox="0 0 22 22"
                        fill="none"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <g opacity="0.5">
                          <path
                            d="M19.2516 3.30005H2.75156C1.58281 3.30005 0.585938 4.26255 0.585938 5.46567V16.6032C0.585938 17.7719 1.54844 18.7688 2.75156 18.7688H19.2516C20.4203 18.7688 21.4172 17.8063 21.4172 16.6032V5.4313C21.4172 4.26255 20.4203 3.30005 19.2516 3.30005ZM19.2516 4.84692C19.2859 4.84692 19.3203 4.84692 19.3547 4.84692L11.0016 10.2094L2.64844 4.84692C2.68281 4.84692 2.71719 4.84692 2.75156 4.84692H19.2516ZM19.2516 17.1532H2.75156C2.40781 17.1532 2.13281 16.8782 2.13281 16.5344V6.35942L10.1766 11.5157C10.4172 11.6875 10.6922 11.7563 10.9672 11.7563C11.2422 11.7563 11.5172 11.6875 11.7578 11.5157L19.8016 6.35942V16.5688C19.8703 16.9125 19.5953 17.1532 19.2516 17.1532Z"
                            fill=""
                          />
                        </g>
                      </svg>
                    </span>
                  </div>
                  {email && (
                    <div className="mt-1 text-right">
                      <button
                        type="button"
                        onClick={handleNotYou}
                        className="text-sm text-primary hover:underline"
                      >
                        Not you?
                      </button>
                    </div>
                  )}
                </div>

                <div className="mb-5">
                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="w-full cursor-pointer rounded-lg border border-primary bg-primary p-4 text-white transition hover:bg-opacity-90 disabled:opacity-50"
                  >
                    {isSubmitting ? 'Sending...' : 'Send Validation Email'}
                  </button>
                </div>

                <div className="mt-6 text-center">
                  <p>
                    Remember your password?{" "}
                    <Link href="/auth" className="text-primary">
                      Sign In
                    </Link>
                  </p>
                </div>

                <div className="mt-2 text-center">
                  <p>
                    Already have a validation code?{" "}
                    <Link href="/auth/validate-email" className="text-primary">
                      Validate Email
                    </Link>
                  </p>
                </div>

                {toast && toast.type === 'success' && (
                  <div className="mt-4 p-4 bg-green-50 dark:bg-green-900/20 rounded-lg">
                    <p className="text-sm text-green-800 dark:text-green-200">
                      <strong>Tip:</strong> If you don't see the email in your inbox, please check your spam or junk folder.
                    </p>
                  </div>
                )}
              </form>

              {/* Debug information section - only visible in development */}
              {isDevelopment && debugInfo && (
                <div className="mt-8 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
                  <h3 className="text-lg font-medium mb-2">Debug Information</h3>
                  <div className="text-xs font-mono overflow-auto">
                    <p><strong>API Endpoint:</strong> {process.env.NEXT_PUBLIC_GRPC_API}</p>
                    <p><strong>Request Time:</strong> {debugInfo.requestTime}</p>
                    <p><strong>Email:</strong> {debugInfo.email}</p>
                    <p><strong>Request Sent:</strong> {debugInfo.requestSent ? 'Yes' : 'No'}</p>
                    <p><strong>Response Status:</strong> {debugInfo.responseStatus}</p>
                    {debugInfo.errorDetails && (
                      <div className="mt-2">
                        <p><strong>Error Details:</strong></p>
                        <pre className="bg-red-50 dark:bg-red-900/20 p-2 rounded mt-1 whitespace-pre-wrap">
                          {debugInfo.errorDetails}
                        </pre>
                      </div>
                    )}
                    <p className="mt-4 text-sm text-gray-600 dark:text-gray-400">
                      Note: Check browser console for additional debug logs.
                    </p>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default ResendValidation;
