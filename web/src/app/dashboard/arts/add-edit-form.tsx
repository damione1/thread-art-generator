"use client";
import { useForm } from "react-hook-form";
import { useRouter } from "next/navigation";

import {  Loader2} from "lucide-react";
import { Art, UpdateArtRequest } from "../../../../grpc/art_pb";
import { useSession } from "next-auth/react";
import { FieldMask } from "../../../../grpc/google/protobuf/field_mask_pb";
import { ArtGeneratorServiceClient } from "../../../../grpc/ServicesServiceClientPb";
import parseError from "../../../../util/errors";
import ImageCropUpload from "./image-upload";

export function AddEditArt({ art }: { art: Art | null }) {
  const { data: session, status, update } = useSession()

  const router = useRouter();
  const isAddMode = !art;
  type FormValues = {
    title: string;
  };
  const defaultValues = isAddMode
    ? {
        title: "",
        name: "",
        imageUrl: "",
        updatedAt: "",
      }
    : {
        title: art.getTitle(),
        name: art.getName(),
        imageUrl: art.getImageUrl(),
        updatedAt: art.getUpdateTime()?.toDate().toISOString(),
      };


  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting, isSubmitSuccessful, isValid },
    reset,
    watch,
    trigger,
    control,
    setValue,
    setFocus,
    setError
  } = useForm<FormValues>({
    defaultValues,
  });

  async function onSubmit(fields: FormValues) {
    update(); // Update the session to get the latest token
    const artPayload = new Art();
    artPayload.setName(art?.getName() ?? "");
    artPayload.setTitle(fields.title);


    let updateMaskPaths = ["title"];
    const updateMask = new FieldMask();
    updateMask.setPathsList(updateMaskPaths);

    const updateArtRequest = new UpdateArtRequest();
    updateArtRequest.setArt(artPayload);
    updateArtRequest.setUpdateMask(updateMask);

    const client = new ArtGeneratorServiceClient(process.env.NEXT_PUBLIC_GRPC_API as string);

    client.updateArt(updateArtRequest, {'Authorization': 'Bearer ' + session?.backendTokens.accessToken}).then((response) => {
     //to implement
    }).catch((error) => {
      console.error("Error updating art", error);
      parseError(error, setError)
    });
  }

  const handleImageChange = async function(croppedImageBlob: Blob)  {
    const formData = new FormData();

    const filename = `${art.getName()}.png`; // Or use any other desired extension
    const file = new File([croppedImageBlob], filename, { type: croppedImageBlob.type });

    formData.append('file', file);
    formData.append('name', art.getName());

    try {
        const response = await fetch(`${process.env.NEXT_PUBLIC_HTTP_API}/v1/upload`, {
            method: "POST",
            body: formData,
            headers: {
                Authorization: `Bearer ${session?.backendTokens?.accessToken}`,
                Accept: "application/json",
            },
        });
        console.log('File uploaded successfully', response);
        return true;
    } catch (error) {
        console.error('Error uploading file:', error);
        return false;
    }
  };



  return (
    <div className="grid grid-cols-5 gap-8">
      <div className="col-span-5 xl:col-span-3">
        <form onSubmit={handleSubmit(onSubmit)}>
          <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
            <div className="flex flex-col gap-5.5 p-6.5">
              <div>
                <label className="mb-3 block text-black dark:text-white">
                  Title
                </label>
                <input
                  type="text"
                  {...register("title")}
                  disabled={isSubmitting}
                  className={`w-full rounded-lg border-[1.5px] border-stroke bg-transparent py-3 px-5 font-medium outline-none transition focus:border-primary active:border-primary disabled:cursor-default disabled:bg-whiter dark:border-form-strokedark dark:bg-form-input dark:focus:border-primary  dark:disabled:bg-black ${
                    errors.title ? "is-invalid" : ""
                  }`}
                />
                <div className="invalid-feedback mt-2">{errors.title?.message}</div>
              </div>

              {errors.root && (
                    <div className="rounded-sm bg-red-50 border border-red-300 text-red-700 p-4 mb-6 mt-5">
                      {errors.root.message}
                    </div>
                  )}
              <div className="grid grid-cols-4 col-span-3 gap-1 mt-4">
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
            </div>
          </div>
        </form>
      </div>
      <div className="col-span-5 xl:col-span-2">
        <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
          <div className="flex flex-col gap-5.5 p-6.5">
            <ImageCropUpload sourceImage={art?.getImageUrl() ?? ''} handleImageChange={handleImageChange} />
          </div>
        </div>
      </div>
    </div>
  );
}
