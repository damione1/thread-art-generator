'use client';

import React, { FormEvent, useState, ChangeEvent } from "react";
import Link from "next/link";
import Breadcrumb from "@/components/Breadcrumbs/Breadcrumb";
import { CreateUserRequest, User } from "@/../grpc/user_pb";
import { ArtGeneratorServiceClient } from "@/../grpc/ServicesServiceClientPb";
import { useRouter } from "next/navigation";
import Toast from "@/components/Notification/Toast";
import { parseValidationErrors } from "@/utils/errorUtils";

const signUpDisabled = false;

const SignUp: React.FC = () => {
  const router = useRouter();
  const [errors, setErrors] = useState<{ [key: string]: string }>({});
  const [toast, setToast] = useState<{ message: string; type: 'error' | 'success' | 'info' } | null>(null);

  // Handle input change to clear error for that field
  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { name } = e.target;
    if (errors[name]) {
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[name];
        return newErrors;
      });
    }
  };

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setErrors({});
    setToast(null);
    const formData = new FormData(e.currentTarget);
    const email = formData.get("email") as string;
    const password = formData.get("password") as string;
    const passwordConfirm = formData.get("passwordConfirm") as string;
    const firstName = formData.get("firstName") as string;
    const lastName = formData.get("lastName") as string;

    // Validate password confirmation
    if (password !== passwordConfirm) {
      setErrors(prev => ({ ...prev, passwordConfirm: "Passwords do not match" }));
      return;
    }

    const client = new ArtGeneratorServiceClient(process.env.NEXT_PUBLIC_GRPC_API!);
    const request = new CreateUserRequest();
    const user = new User();
    user.setEmail(email);
    user.setPassword(password);
    user.setFirstName(firstName);
    user.setLastName(lastName);
    request.setUser(user);

    try {
      const response = await client.createUser(request);
      console.log("User created successfully:", response);
      setToast({
        type: 'success',
        message: 'Account created successfully! Please check your email to validate your account.'
      });

      // Redirect to validate-email page after 2 seconds
      setTimeout(() => {
        router.push(`/auth/validate-email?email=${encodeURIComponent(email)}`);
      }, 2000);
    } catch (error: any) {
      console.error("Error creating user:", error);
      const errorMessage = error.message as string;

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
          message: 'An error occurred while creating your account. Please try again later.'
        });
      }
    }
  }

  return (
    <>
      <Breadcrumb pageName="Sign Up" />

      <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
        <div className="flex flex-wrap items-center">

          <div className="w-full border-stroke dark:border-strokedark xl:w-1/2 xl:border-l-2">
            <div className="w-full p-4 sm:p-12.5 xl:p-17.5">
              {signUpDisabled ? (
                <div className="mb-6">
                  <h2 className="mb-9 text-2xl font-bold text-black dark:text-white sm:text-title-xl2">
                    Sign Up Disabled
                  </h2>
                  <p className="text-black dark:text-white">
                    Sign up is currently disabled. Please contact the administrator.
                  </p>
                </div>
              ) : (
                <>
                  <h2 className="mb-9 text-2xl font-bold text-black dark:text-white sm:text-title-xl2">
                    Sign Up to Thread Art Generator
                  </h2>

                  <form onSubmit={onSubmit}>
                    <div className="mb-4 flex gap-4">
                      <div className="w-1/2">
                        <label className="mb-2.5 block font-medium text-black dark:text-white">
                          First Name
                        </label>
                        <div className="relative">
                          <input
                            type="text"
                            name="firstName"
                            placeholder="Enter your first name"
                            className={`w-full rounded-lg border ${errors.firstName ? 'border-red-500' : 'border-stroke'} bg-transparent py-4 pl-6 pr-10 outline-none focus:border-primary focus-visible:shadow-none dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary`}
                            onChange={handleInputChange}
                          />
                          {errors.firstName && (
                            <span className="text-red-500 text-sm mt-1">{errors.firstName}</span>
                          )}
                        </div>
                      </div>

                      <div className="w-1/2">
                        <label className="mb-2.5 block font-medium text-black dark:text-white">
                          Last Name
                        </label>
                        <div className="relative">
                          <input
                            type="text"
                            name="lastName"
                            placeholder="Enter your last name"
                            className={`w-full rounded-lg border ${errors.lastName ? 'border-red-500' : 'border-stroke'} bg-transparent py-4 pl-6 pr-10 outline-none focus:border-primary focus-visible:shadow-none dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary`}
                            onChange={handleInputChange}
                          />
                          {errors.lastName && (
                            <span className="text-red-500 text-sm mt-1">{errors.lastName}</span>
                          )}
                        </div>
                      </div>
                    </div>

                    <div className="mb-4">
                      <label className="mb-2.5 block font-medium text-black dark:text-white">
                        Email
                      </label>
                      <div className="relative">
                        <input
                          type="email"
                          name="email"
                          placeholder="Enter your email"
                          className={`w-full rounded-lg border ${errors.email ? 'border-red-500' : 'border-stroke'} bg-transparent py-4 pl-6 pr-10 outline-none focus:border-primary focus-visible:shadow-none dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary`}
                          onChange={handleInputChange}
                        />
                        {errors.email && (
                          <span className="text-red-500 text-sm mt-1">{errors.email}</span>
                        )}
                      </div>
                    </div>

                    <div className="mb-4">
                      <label className="mb-2.5 block font-medium text-black dark:text-white">
                        Password
                      </label>
                      <div className="relative">
                        <input
                          type="password"
                          name="password"
                          placeholder="Enter your password"
                          className={`w-full rounded-lg border ${errors.password ? 'border-red-500' : 'border-stroke'} bg-transparent py-4 pl-6 pr-10 outline-none focus:border-primary focus-visible:shadow-none dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary`}
                          onChange={handleInputChange}
                        />
                        {errors.password && (
                          <span className="text-red-500 text-sm mt-1">{errors.password}</span>
                        )}
                      </div>
                    </div>

                    <div className="mb-6">
                      <label className="mb-2.5 block font-medium text-black dark:text-white">
                        Confirm Password
                      </label>
                      <div className="relative">
                        <input
                          type="password"
                          name="passwordConfirm"
                          placeholder="Confirm your password"
                          className={`w-full rounded-lg border ${errors.passwordConfirm ? 'border-red-500' : 'border-stroke'} bg-transparent py-4 pl-6 pr-10 outline-none focus:border-primary focus-visible:shadow-none dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary`}
                          onChange={handleInputChange}
                        />
                        {errors.passwordConfirm && (
                          <span className="text-red-500 text-sm mt-1">{errors.passwordConfirm}</span>
                        )}
                      </div>
                    </div>

                    <div className="mb-5">
                      <input
                        type="submit"
                        value="Create account"
                        className="w-full cursor-pointer rounded-lg border border-primary bg-primary p-4 text-white transition hover:bg-opacity-90"
                      />
                    </div>

                    <div className="mt-6 text-center">
                      <p>
                        Already have an account?{" "}
                        <Link href="/auth" className="text-primary">
                          Sign In
                        </Link>
                      </p>
                    </div>
                  </form>
                </>
              )}
            </div>
          </div>
        </div>
      </div>

      {toast && (
        <Toast
          message={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}
    </>
  );
};

export default SignUp;
