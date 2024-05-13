"use client";
//import ImageUploadForm from "@/components/Images/ImageUploadForm";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { RedirectType, redirect, useRouter } from "next/navigation";
import { revalidatePath } from "next/cache";
import { File, Loader2, Pen, Pickaxe } from "lucide-react";
import { UseGrpcClient } from "@/lib/grpc-context";
import { get } from "http";
import { GetUserRequest, UpdateUserRequest, User } from "../../../../grpc/user_pb";
import { useSession } from "next-auth/react";
import Image from "next/image";
import { access } from "fs";
import { FieldMask } from "../../../../grpc/google/protobuf/field_mask_pb";
import { ArtGeneratorServiceClient } from "../../../../grpc/ServicesServiceClientPb";
import { Header, HeaderParameter } from "../../../../grpc/protoc-gen-openapiv2/options/openapiv2_pb";
import { useEffect, useState } from "react";
//import { redirect } from "next/navigation";

export default function EditProfile({defaultValues}: {defaultValues: any}) {
  const { data: session, status, update } = useSession()

  type FormValues = {
    firstName: string;
    lastName: string;
    email: string;
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
  } = useForm<FormValues>({ defaultValues });

  async function onSubmit(fields: FormValues) {
    console.log("onsubmit fields",fields);

    const user = new User
    user.setName(session?.user.id as unknown as string)
    user.setFirstName(fields.firstName)
    user.setLastName(fields.lastName)

    const updateMask = new FieldMask;
    updateMask.setPathsList(["firstName", "lastName"]);

    const updateUserRequest = new UpdateUserRequest;
    updateUserRequest.setUser(user)
    updateUserRequest.setUpdateMask(updateMask);


    const client = new ArtGeneratorServiceClient("http://localhost:8080");
    const metadata = {
        'Authorization': 'Bearer ' + session?.backendTokens.accessToken
    }

    client.updateUser(updateUserRequest, metadata).then((response) => {
        update({ name: `${fields.firstName} ${fields.lastName}`, email: fields.email });
      console.log("User updated", response);
    }).catch((error) => {
      console.error("Error updating user", error);
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
                  {...register("firstName")}
                  disabled={isSubmitting}
                  className={`w-full rounded-lg border-[1.5px] border-stroke bg-transparent py-3 px-5 font-medium outline-none transition focus:border-primary active:border-primary disabled:cursor-default disabled:bg-whiter dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary  dark:disabled:bg-black ${
                    errors.title ? "is-invalid" : ""
                  }`}
                />
                <div className="invalid-feedback">
                  {/* {errors.firstName?.message} */}
                  </div>
              </div>

              <div>
                <label className="mb-3 block text-black dark:text-white">
                  Last Name
                </label>
                <input
                  type="text"
                  {...register("lastName")}
                  disabled={isSubmitting}
                  className={`w-full rounded-lg border-[1.5px] border-stroke bg-transparent py-3 px-5 font-medium outline-none transition focus:border-primary active:border-primary disabled:cursor-default disabled:bg-whiter dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary dark:disabled:bg-black ${
                    errors.title ? "is-invalid" : ""
                  }`}
                />
                <div className="invalid-feedback">
                  {/* {errors.lastName?.message} */}
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
                <div className="invalid-feedback">
                  {/* {errors.email?.message} */}
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

            </div>
          </div>
          <div className="grid grid-cols-3 gap-3 mt-4">
            <button
              type="submit"
              disabled={isSubmitting}
              className="inline-flex items-center justify-center rounded-md bg-primary px-10 py-4 text-center font-medium text-white hover:bg-opacity-90 lg:px-8 xl:px-10"
            >
              {isSubmitting && (
                <span className="animate-spin mr-4">
                  <Loader2 />
                </span>
              )}
              Save
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
