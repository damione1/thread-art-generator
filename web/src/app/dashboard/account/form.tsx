"use client";

import { useForm } from "react-hook-form";
import Link from "next/link";
import {  Loader2 } from "lucide-react";
import {  UpdateUserRequest, User } from "../../../../grpc/user_pb";
import { useSession } from "next-auth/react";
import Image from "next/image";
import { FieldMask } from "../../../../grpc/google/protobuf/field_mask_pb";
import { ArtGeneratorServiceClient } from "../../../../grpc/ServicesServiceClientPb";
import parseError from "../../../../util/errors";

export default function EditProfile({defaultValues}: {defaultValues: any}) {
  const { data: session, status, update } = useSession()

  type FormValues = {
    first_name: string;
    last_name: string;
    email: string;
    password: string;
    confirmPassword: string;
  };

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting, isSubmitSuccessful },
    reset,
    watch,
    trigger,
    control,
    setValue,
    setFocus,
    setError
  } = useForm<FormValues>({ defaultValues });

  async function onSubmit(fields: FormValues) {
    update(); // Update the session to get the latest token
    const user = new User();
    user.setName(session?.user.id as unknown as string);
    user.setFirstName(fields.first_name);
    user.setLastName(fields.last_name);


    let updateMaskPaths = ["first_name", "last_name"];
    if (fields.password) {
      updateMaskPaths.push("password");
      user.setPassword(fields.password);
    }
    const updateMask = new FieldMask();
    updateMask.setPathsList(updateMaskPaths);

    const updateUserRequest = new UpdateUserRequest();
    updateUserRequest.setUser(user);
    updateUserRequest.setUpdateMask(updateMask);

    const client = new ArtGeneratorServiceClient(process.env.NEXT_PUBLIC_GRPC_API as string);

    client.updateUser(updateUserRequest, {'Authorization': 'Bearer ' + session?.backendTokens.accessToken}).then((response) => {
      // Update the session with the new user details
      update({user:{
        name: `${response.getFirstName()} ${response.getLastName()}`,
        email: response.getEmail(),
        image: response.getAvatar()
      }});
    }).catch((error) => {
      console.error("Error updating user", error);
      parseError(error, setError)
    });
  }


  return (
    <div className="grid grid-cols-5 gap-8">
      <div className="col-span-5 xl:col-span-3">
        <form onSubmit={handleSubmit(onSubmit)}>
          <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
            <div className="flex flex-col gap-5.5 p-6.5">

              <div>
                <label className="mb-3 block text-black dark:text-white">
                  First Name
                </label>
                <input
                  type="text"
                  {...register("first_name")}
                  disabled={isSubmitting}
                  className={`w-full rounded-lg border-[1.5px] border-stroke bg-transparent py-3 px-5 font-medium outline-none transition focus:border-primary active:border-primary disabled:cursor-default disabled:bg-whiter dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary  dark:disabled:bg-black ${
                    errors.first_name ? "is-invalid" : ""
                  }`}
                />
                <div className="invalid-feedback mt-2">
                   {errors.first_name?.message}
                  </div>
              </div>

              <div>
                <label className="mb-3 block text-black dark:text-white">
                  Last Name
                </label>
                <input
                  type="text"
                  {...register("last_name")}
                  disabled={isSubmitting}
                  className={`w-full rounded-lg border-[1.5px] border-stroke bg-transparent py-3 px-5 font-medium outline-none transition focus:border-primary active:border-primary disabled:cursor-default disabled:bg-whiter dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary dark:disabled:bg-black ${
                    errors.last_name ? "is-invalid" : ""
                  }`}
                />
                <div className="invalid-feedback mt-2">
                 {errors.last_name?.message}
                </div>
              </div>

              <div>
                <label className="mb-3 block text-black dark:text-white">
                  Email
                </label>
                <input
                  disabled
                  {...register("email")}
                  className="w-full rounded-lg border-[1.5px] border-stroke bg-transparent py-3 px-5 font-medium outline-none transition focus:border-primary active:border-primary disabled:cursor-default disabled:bg-whiter dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary dark:disabled:bg-black"
                ></input>
                <div className="invalid-feedback mt-2">
                 {errors.email?.message}
                </div>
              </div>

              <div>
                <label className="mb-3 block text-black dark:text-white">
                  Profile Image
                </label>
                <div className="relative z-20">
                <span className="h-12 w-12 rounded-full overflow-hidden">
                <p className="text-sm text-gray-500 mt-0">
                  You can change your profile avatar on <Link href="https://gravatar.com" target="_blank" className="text-primary">Gravatar</Link>
                </p>
                  <Image
                    width={112}
                    height={112}
                    src={session?.user.image as string || "/images/avatar.png"}
                    alt="User"
                  />
                </span>
                </div>
              </div>

              <div>
                <label className="mb-3 block text-black dark:text-white">
                  Password
                </label>
                <input
                  type="password"
                  {...register("password")}
                  className="w-full rounded-lg border-[1.5px] border-stroke bg-transparent py-3 px-5 font-medium outline-none transition focus:border-primary active:border-primary disabled:cursor-default disabled:bg-whiter dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary dark:disabled:bg-black"
                ></input>
                <div className="invalid-feedback mt-2">
                  {errors.password?.message}
                </div>
              </div>

              <div>
                <label className="mb-3 block text-black dark:text-white">
                  Confirm Password
                </label>
                <input
                  type="password"
                  {...register("confirmPassword", {
                    validate: value =>
                      value === watch('password') || "The passwords do not match"
                  })}
                  className="w-full rounded-lg border-[1.5px] border-stroke bg-transparent py-3 px-5 font-medium outline-none transition focus:border-primary active:border-primary disabled:cursor-default disabled:bg-whiter dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary dark:disabled:bg-black"
                ></input>
                <div className="invalid-feedback mt-2">
                  {errors.confirmPassword?.message}
                </div>
              </div>

            </div>
          </div>
            {errors.root && (
                <div className="rounded-sm bg-red-50 border border-red-300 text-red-700 p-4 mb-6 mt-5">
                  {errors.root.message}
                </div>
              )}
          <div className="grid grid-cols-3 gap-3 mt-4">
            <button
              type="submit"
              disabled={isSubmitting}
              className="inline-flex items-center justify-center rounded-md bg-primary px-10 py-4 text-center font-medium text-white hover:bg-opacity-90 lg:px-8 xl:px-10"
            >
              {isSubmitting && (
                <Loader2 className="animate-spin mr-4" />
              )}
              Save
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
