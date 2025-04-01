"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { useParams } from "next/navigation";
import Image from "next/image";
import {
  getArt,
  getArtUploadUrl,
  confirmArtImageUpload,
  getCurrentUser,
  deleteArt,
} from "@/lib/grpc-client";
import { Art } from "@/lib/pb/art_pb";
import { ErrorMessage, SuccessMessage } from "@/components/ui";
import { Cropper, CropperRef, CircleStencil } from "react-advanced-cropper";
import "react-advanced-cropper/dist/style.css";
import { getStatusInfo } from "@/utils/artUtils";

export default function ArtDetailPage() {
  const router = useRouter();
  const params = useParams();
  const artId = params?.id as string;
  const cropperRef = useRef<CropperRef>(null);

  const [art, setArt] = useState<Art | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Upload states
  const [isUploading, setIsUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [uploadSuccess, setUploadSuccess] = useState(false);

  // Cropper states
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [cropperImage, setCropperImage] = useState<string | null>(null);
  const [showCropper, setShowCropper] = useState(false);

  //States for delete confirmation
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  // Fetch art and user data
  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true);
      setError(null);
      try {
        // Get current user
        const currentUser = await getCurrentUser();

        // Get art details
        if (artId) {
          // Format the art resource name correctly
          const artResourceName = `${currentUser.name}/arts/${artId}`;
          const artData = await getArt(artResourceName);
          setArt(artData);
        }
      } catch (err) {
        console.error("Error fetching data:", err);
        setError("Failed to load art details. Please try again.");
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [artId]);

  // Handle file upload
  const handleFileUpload = async (file: File) => {
    if (!art || !art.name) {
      setUploadError("Art information not available. Please try again.");
      return;
    }

    setIsUploading(true);
    setUploadError(null);
    setUploadSuccess(false);

    try {
      // Get the cropped canvas and create a blob from it
      const canvas = cropperRef.current?.getCanvas();
      if (!canvas) {
        throw new Error("Failed to get cropped image");
      }

      // Convert canvas to blob
      const blob = await new Promise<Blob>((resolve, reject) => {
        canvas.toBlob((blob) => {
          if (blob) resolve(blob);
          else reject(new Error("Failed to create image blob"));
        }, file.type);
      });

      // Create File object from Blob
      const croppedFile = new File([blob], file.name, { type: file.type });

      // Get upload URL
      const uploadUrlResponse = await getArtUploadUrl(art.name);

      // Upload file to the signed URL
      const response = await fetch(uploadUrlResponse.uploadUrl, {
        method: "PUT",
        body: croppedFile,
        headers: {
          "Content-Type": file.type,
          "x-amz-acl": "private",
        },
        mode: "cors",
        credentials: "omit",
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

      // Reset cropper state
      setSelectedFile(null);
      setCropperImage(null);
      setShowCropper(false);
      setUploadSuccess(true);

      // Reload art data after a short delay
      setTimeout(() => {
        router.refresh();
      }, 1500);
    } catch (error) {
      console.error("Upload failed:", error);
      setUploadError("Failed to upload image. Please try again.");
    } finally {
      setIsUploading(false);
    }
  };

  // Process the selected file and prepare for cropping
  const processFile = (file: File) => {
    if (!file.type.startsWith("image/")) {
      setUploadError("Please upload an image file (JPEG, PNG, etc.)");
      return;
    }

    setSelectedFile(file);
    const reader = new FileReader();
    reader.onload = (e) => {
      if (e.target?.result) {
        setCropperImage(e.target.result as string);
        setShowCropper(true);
      }
    };
    reader.readAsDataURL(file);
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
      processFile(file);
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      const file = e.target.files[0];
      processFile(file);
    }
  };

  // Cancel cropping and reset state
  const handleCancelCrop = () => {
    setSelectedFile(null);
    setCropperImage(null);
    setShowCropper(false);
  };

  // Submit cropped image
  const handleCropSubmit = () => {
    if (selectedFile) {
      handleFileUpload(selectedFile);
    }
  };

  // Handle art deletion
  const handleDeleteArt = async () => {
    if (!art || !art.name) {
      return;
    }

    setIsDeleting(true);
    setDeleteError(null);

    try {
      await deleteArt(art.name);

      // Redirect to dashboard after successful deletion
      router.push("/dashboard");
    } catch (error) {
      console.error("Failed to delete art:", error);
      setDeleteError("Failed to delete art. Please try again.");
      setIsDeleting(false);
      setShowDeleteConfirm(false);
    }
  };

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-12">
        <div className="max-w-4xl mx-auto">
          <div className="flex justify-center items-center min-h-[300px]">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary-500"></div>
          </div>
        </div>
      </div>
    );
  }

  if (error || !art) {
    return (
      <div className="container mx-auto px-4 py-12">
        <div className="max-w-4xl mx-auto">
          <div className="bg-dark-200 rounded-lg p-6 shadow-lg">
            <h2 className="text-xl font-semibold mb-4 text-slate-100">Error</h2>
            <p className="text-slate-300 mb-4">{error || "Art not found"}</p>
            <button
              onClick={() => router.push("/dashboard")}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
            >
              Back to Dashboard
            </button>
          </div>
        </div>
      </div>
    );
  }

  const statusInfo = getStatusInfo(art.status);
  const needsImage = art.status === 1; // ART_STATUS_PENDING_IMAGE
  const created = art.createTime
    ? new Date(Number(art.createTime.seconds) * 1000)
    : new Date();

  return (
    <div className="container mx-auto px-4 py-12">
      <div className="max-w-4xl mx-auto">
        {/* Delete confirmation modal */}
        {showDeleteConfirm && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-dark-200 rounded-lg p-6 w-full max-w-md">
              <h3 className="text-xl font-semibold mb-4 text-slate-100">
                Delete Art
              </h3>
              <p className="text-slate-300 mb-6">
                Are you sure you want to delete &ldquo;{art.title}&rdquo;? This
                action cannot be undone.
              </p>
              {deleteError && (
                <div className="mb-4 p-3 bg-red-900/30 border border-red-700 text-red-400 rounded">
                  {deleteError}
                </div>
              )}
              <div className="flex justify-end space-x-3">
                <button
                  onClick={() => setShowDeleteConfirm(false)}
                  disabled={isDeleting}
                  className="px-4 py-2 bg-dark-300 text-slate-300 rounded hover:bg-dark-400 transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleDeleteArt}
                  disabled={isDeleting}
                  className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition-colors flex items-center"
                >
                  {isDeleting ? (
                    <>
                      <svg
                        className="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                      >
                        <circle
                          className="opacity-25"
                          cx="12"
                          cy="12"
                          r="10"
                          stroke="currentColor"
                          strokeWidth="4"
                        ></circle>
                        <path
                          className="opacity-75"
                          fill="currentColor"
                          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                        ></path>
                      </svg>
                      Deleting...
                    </>
                  ) : (
                    "Delete"
                  )}
                </button>
              </div>
            </div>
          </div>
        )}

        <div className="bg-dark-200 rounded-lg p-6 shadow-lg mb-6">
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-2xl font-semibold text-slate-100">
              {art.title}
            </h2>
            <span
              className={`${statusInfo.color} px-3 py-1 rounded-full text-sm border border-current`}
            >
              {statusInfo.text}
            </span>
          </div>

          <div className="mb-6">
            <p className="text-slate-400 text-sm">
              Created on {created.toLocaleDateString()} at{" "}
              {created.toLocaleTimeString()}
            </p>
          </div>

          {needsImage ? (
            <div className="mb-6">
              <ErrorMessage message={uploadError} />
              <SuccessMessage
                message={uploadSuccess ? "Image uploaded successfully!" : null}
              />

              {showCropper && cropperImage ? (
                <div>
                  <div className="rounded-lg overflow-hidden bg-dark-300 p-4 mb-4">
                    <Cropper
                      ref={cropperRef}
                      src={cropperImage}
                      className="h-[500px] rounded"
                      stencilComponent={CircleStencil}
                      stencilProps={{ aspectRatio: 1 }}
                    />
                  </div>
                  <div className="flex justify-between">
                    <button
                      onClick={handleCancelCrop}
                      className="px-4 py-2 bg-dark-300 text-slate-300 rounded hover:bg-dark-400 transition-colors"
                      disabled={isUploading}
                    >
                      Cancel
                    </button>
                    <button
                      onClick={handleCropSubmit}
                      className="px-4 py-2 bg-primary-500 text-white rounded hover:bg-primary-600 transition-colors flex items-center"
                      disabled={isUploading}
                    >
                      {isUploading ? (
                        <>
                          <svg
                            className="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                          >
                            <circle
                              className="opacity-25"
                              cx="12"
                              cy="12"
                              r="10"
                              stroke="currentColor"
                              strokeWidth="4"
                            ></circle>
                            <path
                              className="opacity-75"
                              fill="currentColor"
                              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                            ></path>
                          </svg>
                          Uploading...
                        </>
                      ) : (
                        "Upload Image"
                      )}
                    </button>
                  </div>
                </div>
              ) : (
                <div
                  className={`border-2 border-dashed rounded-lg p-8 mb-4 text-center cursor-pointer ${
                    isUploading
                      ? "bg-dark-300 border-gray-600"
                      : "border-primary-400 hover:border-primary-300"
                  }`}
                  onDragOver={handleDragOver}
                  onDrop={handleDrop}
                  onClick={() =>
                    document.getElementById("file-upload")?.click()
                  }
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
                    <p className="text-slate-400 text-sm">
                      JPEG, PNG (max 10MB)
                    </p>

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
              )}
            </div>
          ) : art.imageUrl ? (
            <div className="mb-6">
              <div className="rounded-lg overflow-hidden bg-dark-300 aspect-square w-full max-w-xl mx-auto">
                <Image
                  src={art.imageUrl}
                  alt={art.title}
                  className="w-full h-full object-contain"
                  width={600}
                  height={600}
                />
              </div>
            </div>
          ) : (
            <div className="mb-6 p-4 bg-dark-300 rounded-lg text-center">
              <p className="text-slate-300">No image available</p>
            </div>
          )}

          <div className="flex justify-between mt-6">
            <button
              onClick={() => router.push("/dashboard")}
              className="px-4 py-2 bg-dark-300 text-slate-300 rounded hover:bg-dark-400 transition-colors"
            >
              Back to Dashboard
            </button>

            <button
              onClick={() => setShowDeleteConfirm(true)}
              className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition-colors"
            >
              Delete Art
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
