'use client';

import React, { useState, ChangeEvent, FormEvent, useEffect } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import Breadcrumb from "@/components/Breadcrumbs/Breadcrumb";
import { ValidateEmailRequest } from "@/../grpc/user_pb";
import { ArtGeneratorServiceClient } from "@/../grpc/ServicesServiceClientPb";
import Toast from "@/components/Notification/Toast";
import { parseValidationErrors } from "@/utils/errorUtils";
import { saveEmail, getSavedEmail, clearSavedEmail } from "@/utils/authUtils";

const ValidateEmail: React.FC = () => {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [errors, setErrors] = useState<{ [key: string]: string }>({});
  const [toast, setToast] = useState<{ message: string; type: 'error' | 'success' | 'info' } | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formData, setFormData] = useState({
    email: '',
    validationNumber: ''
  });
  const [debugInfo, setDebugInfo] = useState<{
    requestSent: boolean;
    requestTime: string;
    email: string;
    validationNumber: string;
    responseStatus: string;
    errorDetails: string;
  } | null>(null);

  // Initialize form data with URL parameters or saved email
  useEffect(() => {
    const urlEmail = searchParams?.get('email') || '';
    const urlCode = searchParams?.get('code') || '';
    const savedEmail = getSavedEmail();

    setFormData({
      email: urlEmail || savedEmail,
      validationNumber: urlCode
    });
  }, [searchParams]);

  // Handle input change to clear error for that field and update form data
  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;

    // Create a new object instead of mutating the existing one
    const newFormData = { ...formData };
    newFormData[name as keyof typeof formData] = value;
    setFormData(newFormData);

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
    setFormData(prev => ({
      ...prev,
      email: ''
    }));
  };

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setErrors({});
    setToast(null);
    setIsSubmitting(true);
    setDebugInfo(null);

    const { email, validationNumber } = formData;

    // Save email for future use
    saveEmail(email);

    // Initialize debug info
    const debugData = {
      requestSent: false,
      requestTime: new Date().toISOString(),
      email: email,
      validationNumber: validationNumber,
      responseStatus: 'pending',
      errorDetails: ''
    };

    const client = new ArtGeneratorServiceClient(process.env.NEXT_PUBLIC_GRPC_API!);
    const request = new ValidateEmailRequest();
    request.setEmail(email);
    request.setValidationnumber(parseInt(validationNumber, 10));

    try {
      console.log(`[DEBUG] Validating email: ${email} with code: ${validationNumber}`);
      console.log(`[DEBUG] API endpoint: ${process.env.NEXT_PUBLIC_GRPC_API}`);

      debugData.requestSent = true;

      const response = await client.validateEmail(request);

      console.log(`[DEBUG] Email validated successfully, response:`, response);
      debugData.responseStatus = 'success';

      setToast({
        type: 'success',
        message: 'Your email has been successfully validated. You can now log in to your account.'
      });

      // Redirect to login page after 3 seconds
      setTimeout(() => {
        router.push('/auth');
      }, 3000);
    } catch (error: any) {
      console.error("[DEBUG] Error validating email:", error);
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
          // Map the errors to user-friendly messages
          const mappedErrors: { [key: string]: string } = {};

          Object.keys(validationErrors).forEach(key => {
            // Map technical error messages to user-friendly messages
            if (validationErrors[key] === 'account activation not found') {
              mappedErrors[key] = key === 'email'
                ? 'We couldn\'t find an account with this email and validation code'
                : 'Please check your validation code';
            } else {
              mappedErrors[key] = validationErrors[key];
            }
          });

          setErrors(mappedErrors);
        }
      } else {
        // Handle non-field specific errors
        setToast({
          type: 'error',
          message: 'An error occurred while validating your email. Please check your email and validation code and try again.'
        });
      }
    } finally {
      setIsSubmitting(false);
      setDebugInfo(debugData);
    }
  }

  // Determine if we're in development mode
  const isDevelopment = process.env.NODE_ENV === 'development';

  // Auto-submit if both email and validation code are provided in URL
  useEffect(() => {
    if (formData.email && formData.validationNumber && !isSubmitting && !toast) {
      const form = document.getElementById('validate-email-form') as HTMLFormElement;
      if (form) {
        form.dispatchEvent(new Event('submit', { cancelable: true }));
      }
    }
  }, [formData, isSubmitting, toast]);

  return (
    <>
      <Breadcrumb pageName="Validate Email" />

      <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
        <div className="flex flex-wrap items-center">
          <div className="w-full border-stroke dark:border-strokedark xl:w-1/2 xl:border-l-2">
            <div className="w-full p-4 sm:p-12.5 xl:p-17.5">
              <h2 className="mb-9 text-2xl font-bold text-black dark:text-white sm:text-title-xl2">
                Validate Email
              </h2>
              <p className="mb-6 text-base text-body-color dark:text-body-color-dark">
                Enter your email address and validation code to activate your account.
              </p>

              {toast && (
                <Toast
                  message={toast.message}
                  type={toast.type}
                  onClose={() => setToast(null)}
                />
              )}

              <form id="validate-email-form" onSubmit={onSubmit}>
                <div className="mb-4">
                  <label className="mb-2.5 block font-medium text-black dark:text-white">
                    Email
                  </label>
                  <div className="relative">
                    <input
                      type="email"
                      name="email"
                      value={formData.email}
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
                  {formData.email && (
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

                <div className="mb-6">
                  <label className="mb-2.5 block font-medium text-black dark:text-white">
                    Validation Code
                  </label>
                  <div className="relative">
                    <input
                      type="text"
                      name="validationNumber"
                      value={formData.validationNumber}
                      placeholder="Enter your 7-digit validation code"
                      className={`w-full rounded-lg border ${errors.validationNumber ? 'border-red-500' : 'border-stroke'} bg-transparent py-4 pl-6 pr-10 outline-none focus:border-primary focus-visible:shadow-none dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary`}
                      onChange={handleInputChange}
                      required
                      pattern="[0-9]{7}"
                      title="Please enter a 7-digit number"
                    />
                    {errors.validationNumber && (
                      <span className="text-red-500 text-sm mt-1">{errors.validationNumber}</span>
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
                            d="M16.1547 6.80626V5.91251C16.1547 3.16251 14.0922 0.825009 11.4797 0.618759C10.0359 0.481259 8.59219 0.996884 7.52656 1.95938C6.46094 2.92188 5.84219 4.29688 5.84219 5.70626V6.80626C3.84844 7.18438 2.33594 8.93751 2.33594 11.0688V17.2906C2.33594 19.5594 4.19219 21.3813 6.42656 21.3813H15.5016C17.7703 21.3813 19.6266 19.525 19.6266 17.2563V11C19.6609 8.93751 18.1484 7.21876 16.1547 6.80626ZM8.55781 3.09376C9.31406 2.40626 10.3109 2.06251 11.3422 2.16563C13.1641 2.33751 14.6078 3.98751 14.6078 5.91251V6.70313H7.38906V5.67188C7.38906 4.70938 7.80156 3.78126 8.55781 3.09376ZM18.1141 17.2906C18.1141 18.7 16.9453 19.8688 15.5359 19.8688H6.46094C5.05156 19.8688 3.91719 18.7344 3.91719 17.325V11.0688C3.91719 9.52189 5.15469 8.28438 6.70156 8.28438H15.2953C16.8422 8.28438 18.1141 9.52188 18.1141 11V17.2906Z"
                            fill=""
                          />
                          <path
                            d="M10.9977 11.8594C10.5852 11.8594 10.207 12.2031 10.207 12.65V16.2594C10.207 16.6719 10.5508 17.05 10.9977 17.05C11.4102 17.05 11.7883 16.7063 11.7883 16.2594V12.6156C11.7883 12.2031 11.4102 11.8594 10.9977 11.8594Z"
                            fill=""
                          />
                        </g>
                      </svg>
                    </span>
                  </div>
                </div>

                <div className="mb-5">
                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="w-full cursor-pointer rounded-lg border border-primary bg-primary p-4 text-white transition hover:bg-opacity-90 disabled:opacity-50"
                  >
                    {isSubmitting ? 'Validating...' : 'Validate Email'}
                  </button>
                </div>

                <div className="mt-6 text-center">
                  <p>
                    Didn't receive a validation code?{" "}
                    <Link href="/auth/resend-validation" className="text-primary">
                      Send Validation Email
                    </Link>
                  </p>
                </div>
              </form>

              {/* Debug information section - only visible in development */}
              {isDevelopment && debugInfo && (
                <div className="mt-8 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
                  <h3 className="text-lg font-medium mb-2">Debug Information</h3>
                  <div className="text-xs font-mono overflow-auto">
                    <p><strong>API Endpoint:</strong> {process.env.NEXT_PUBLIC_GRPC_API}</p>
                    <p><strong>Request Time:</strong> {debugInfo.requestTime}</p>
                    <p><strong>Email:</strong> {debugInfo.email}</p>
                    <p><strong>Validation Number:</strong> {debugInfo.validationNumber}</p>
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

export default ValidateEmail;
