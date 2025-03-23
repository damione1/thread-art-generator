"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import {
  createArt,
  getArtUploadUrl,
  confirmArtImageUpload,
  getCurrentUser,
} from "@/lib/grpc-client";
import { Art } from "@/lib/pb/art_pb";
import { User } from "@/lib/pb/user_pb";
import {
  FormField,
  TextInput,
  ErrorMessage,
  SuccessMessage,
  useForm,
} from "@/components/ui";

interface ArtFormValues extends Record<string, unknown> {
  title: string;
}

export default function NewArtPage() {
  const router = useRouter();
  const [currentStep, setCurrentStep] = useState<"create" | "upload">("create");
  const [art, setArt] = useState<Art | null>(null);
  const [user, setUser] = useState<User | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [uploadSuccess, setUploadSuccess] = useState(false);

  // Initialize form for art creation
  const {
    values,
    errors,
    handleChange,
    handleSubmit,
    isSubmitting,
    generalError,
    isSuccess,
  } = useForm<ArtFormValues>({
    title: "",
  });

  // Fetch current user if not already loaded
  const fetchCurrentUser = async () => {
    if (!user) {
      try {
        const currentUser = await getCurrentUser();
        setUser(currentUser);
      } catch (error) {
        console.error("Failed to fetch user:", error);
      }
    }
  };

  // Handle form submission to create art
  const handleCreateArt = async (formValues: ArtFormValues) => {
    try {
      await fetchCurrentUser();

      if (!user) {
        throw new Error("User not authenticated");
      }

      const newArt = await createArt({ title: formValues.title }, user.name);

      setArt(newArt);
      setCurrentStep("upload");
    } catch (error) {
      console.error("Failed to create art:", error);
      throw error; // Let the form handler deal with this
    }
  };

  // Handle file drop/selection
  const handleFileUpload = async (file: File) => {
    if (!art || !art.name) {
      setUploadError("No art created yet. Please create art first.");
      return;
    }

    setIsUploading(true);
    setUploadError(null);

    try {
      // Get upload URL
      const uploadUrlResponse = await getArtUploadUrl(art.name);

      // Upload file to the signed URL
      const response = await fetch(uploadUrlResponse.uploadUrl, {
        method: "PUT",
        body: file,
        headers: {
          "Content-Type": file.type,
        },
      });

      if (!response.ok) {
        throw new Error(`Upload failed: ${response.statusText}`);
      }

      // Confirm the upload to update the status to complete
      try {
        const updatedArt = await confirmArtImageUpload(art.name);
        setArt(updatedArt);
      } catch (confirmError) {
        console.error("Failed to confirm upload:", confirmError);
        // We still continue since the image was uploaded successfully
      }

      setUploadSuccess(true);

      // Redirect to art details page after short delay
      setTimeout(() => {
        // Extract the art ID from the full resource name
        const artIdMatch = art.name.match(/users\/[^/]+\/arts\/([^/]+)$/);
        const artId = artIdMatch ? artIdMatch[1] : art.name;

        router.push(`/dashboard/arts/${artId}`);
      }, 1500);
    } catch (error) {
      console.error("Upload failed:", error);
      setUploadError("Failed to upload image. Please try again.");
    } finally {
      setIsUploading(false);
    }
  };

  // Handle file drag and drop
  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      const file = e.dataTransfer.files[0];
      if (file.type.startsWith("image/")) {
        handleFileUpload(file);
      } else {
        setUploadError("Please upload an image file (JPEG, PNG, etc.)");
      }
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      const file = e.target.files[0];
      if (file.type.startsWith("image/")) {
        handleFileUpload(file);
      } else {
        setUploadError("Please upload an image file (JPEG, PNG, etc.)");
      }
    }
  };

  // Initialize component
  useEffect(() => {
    fetchCurrentUser();
  }, []);

  return (
    <div className="container mx-auto px-4 py-12">
      <div className="max-w-2xl mx-auto">
        <div className="bg-dark-200 rounded-lg p-6 shadow-lg">
          <h2 className="text-xl font-semibold mb-4 text-slate-100">
            {currentStep === "create" ? "Create New Art" : "Upload Image"}
          </h2>

          {currentStep === "create" ? (
            <>
              <ErrorMessage message={generalError} />
              <SuccessMessage
                message={
                  isSuccess
                    ? "Art created successfully! Please upload an image."
                    : null
                }
              />

              <form onSubmit={handleSubmit(handleCreateArt)}>
                <div className="mb-4">
                  <FormField
                    id="title"
                    name="title"
                    label="Title"
                    error={errors.title}
                    inputComponent={
                      <TextInput
                        id="title"
                        name="title"
                        value={values.title}
                        onChange={handleChange}
                        disabled={isSubmitting}
                        placeholder="Enter art title"
                      />
                    }
                  />
                </div>

                <div className="flex justify-end">
                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
                  >
                    {isSubmitting ? "Creating..." : "Create Art"}
                  </button>
                </div>
              </form>
            </>
          ) : (
            <>
              <ErrorMessage message={uploadError} />
              <SuccessMessage
                message={uploadSuccess ? "Image uploaded successfully!" : null}
              />

              <div
                className={`border-2 border-dashed rounded-lg p-8 mb-4 text-center cursor-pointer ${
                  isUploading
                    ? "bg-dark-300 border-gray-600"
                    : "border-primary-400 hover:border-primary-300"
                }`}
                onDragOver={handleDragOver}
                onDrop={handleDrop}
                onClick={() => document.getElementById("file-upload")?.click()}
              >
                <input
                  type="file"
                  id="file-upload"
                  className="hidden"
                  accept="image/*"
                  onChange={handleFileSelect}
                  disabled={isUploading}
                />
                <div className="flex flex-col items-center justify-center">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-12 w-12 text-primary-400 mb-4"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                    />
                  </svg>
                  <p className="text-slate-300 mb-2">
                    Drag and drop an image here, or click to select
                  </p>
                  <p className="text-slate-400 text-sm">JPEG, PNG (max 10MB)</p>

                  {isUploading && (
                    <div className="mt-4">
                      <div className="w-full bg-dark-400 rounded-full h-2 mt-2">
                        <div
                          className="bg-primary-500 h-2 rounded-full animate-pulse"
                          style={{ width: "100%" }}
                        ></div>
                      </div>
                      <p className="text-slate-400 text-sm mt-2">
                        Uploading...
                      </p>
                    </div>
                  )}
                </div>
              </div>

              <div className="flex justify-between">
                <button
                  onClick={() => setCurrentStep("create")}
                  disabled={isUploading}
                  className="px-4 py-2 bg-dark-300 text-slate-300 rounded hover:bg-dark-400 transition-colors"
                >
                  Back
                </button>
                <button
                  onClick={() => router.push("/dashboard")}
                  disabled={isUploading}
                  className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
                >
                  Done
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
