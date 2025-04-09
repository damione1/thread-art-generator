"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { useParams } from "next/navigation";
import {
  getArt,
  getArtUploadUrl,
  confirmArtImageUpload,
  getCurrentUser,
  deleteArt,
  createComposition,
  listCompositions,
} from "@/lib/grpc-client";
import { Art, Composition } from "@/lib/pb/art_pb";
import { ErrorMessage, SuccessMessage } from "@/components/ui";
import { Cropper, CropperRef, CircleStencil } from "react-advanced-cropper";
import "react-advanced-cropper/dist/style.css";
import { getStatusInfo } from "@/utils/artUtils";
import Image from "next/image";

interface CompositionFormData {
  nailsQuantity: number;
  imgSize: number;
  maxPaths: number;
  startingNail: number;
  minimumDifference: number;
  brightnessFactor: number;
  imageContrast: number;
  physicalRadius: number;
}

interface CompositionFormProps {
  onSubmit: (formData: CompositionFormData) => void;
  initialValues?: Partial<CompositionFormData> | null;
  isLoading?: boolean;
}

const SliderInput = ({
  label,
  value,
  onChange,
  min,
  max,
  step = 1,
  unit = "",
}: {
  label: string;
  value: number;
  onChange: (value: number) => void;
  min: number;
  max: number;
  step?: number;
  unit?: string;
}) => (
  <div>
    <div className="flex justify-between">
      <label className="block text-sm font-medium text-slate-300">
        {label}
      </label>
      <span className="text-sm text-slate-400">
        {value}
        {unit}
      </span>
    </div>
    <input
      type="range"
      value={value}
      onChange={(e) => onChange(parseFloat(e.target.value))}
      className="mt-2 w-full h-2 bg-dark-400 rounded-lg appearance-none cursor-pointer accent-primary-500"
      min={min}
      max={max}
      step={step}
    />
    <div className="flex justify-between mt-1">
      <span className="text-xs text-slate-500">
        {min}
        {unit}
      </span>
      <span className="text-xs text-slate-500">
        {max}
        {unit}
      </span>
    </div>
  </div>
);

// Composition form component
const CompositionForm = ({
  onSubmit,
  initialValues = null,
  isLoading = false,
}: CompositionFormProps) => {
  const [formData, setFormData] = useState<CompositionFormData>({
    nailsQuantity: initialValues?.nailsQuantity ?? 200,
    imgSize: initialValues?.imgSize ?? 800,
    maxPaths: initialValues?.maxPaths ?? 2000,
    startingNail: initialValues?.startingNail ?? 0,
    minimumDifference: initialValues?.minimumDifference ?? 40,
    brightnessFactor: initialValues?.brightnessFactor ?? 70,
    imageContrast: initialValues?.imageContrast ?? 50,
    physicalRadius: initialValues?.physicalRadius ?? 150,
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="space-y-6 bg-dark-300 p-6 rounded-lg"
    >
      <h3 className="text-xl font-semibold text-slate-100 mb-4">
        New Composition
      </h3>

      <div className="grid grid-cols-1 gap-6">
        <SliderInput
          label="Nails Quantity"
          value={formData.nailsQuantity}
          onChange={(value) =>
            setFormData({ ...formData, nailsQuantity: value })
          }
          min={1}
          max={1000}
        />

        <SliderInput
          label="Image Size"
          value={formData.imgSize}
          onChange={(value) => setFormData({ ...formData, imgSize: value })}
          min={100}
          max={5000}
          step={100}
          unit="px"
        />

        <SliderInput
          label="Max Paths"
          value={formData.maxPaths}
          onChange={(value) => setFormData({ ...formData, maxPaths: value })}
          min={100}
          max={20000}
          step={100}
        />

        <SliderInput
          label="Starting Nail"
          value={formData.startingNail}
          onChange={(value) =>
            setFormData({ ...formData, startingNail: value })
          }
          min={0}
          max={formData.nailsQuantity - 1}
        />

        <SliderInput
          label="Minimum Difference"
          value={formData.minimumDifference}
          onChange={(value) =>
            setFormData({ ...formData, minimumDifference: value })
          }
          min={1}
          max={200}
        />

        <SliderInput
          label="Brightness Factor"
          value={formData.brightnessFactor}
          onChange={(value) =>
            setFormData({ ...formData, brightnessFactor: value })
          }
          min={1}
          max={255}
        />

        <SliderInput
          label="Image Contrast"
          value={formData.imageContrast}
          onChange={(value) =>
            setFormData({ ...formData, imageContrast: value })
          }
          min={0}
          max={100}
          step={0.1}
          unit="%"
        />

        <SliderInput
          label="Physical Radius"
          value={formData.physicalRadius}
          onChange={(value) =>
            setFormData({ ...formData, physicalRadius: value })
          }
          min={50}
          max={500}
          step={1}
          unit="mm"
        />
      </div>

      <div className="flex justify-end mt-6">
        <button
          type="submit"
          disabled={isLoading}
          className="px-4 py-2 bg-primary-500 text-white rounded hover:bg-primary-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center"
        >
          {isLoading ? (
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
              Creating...
            </>
          ) : (
            "Create Composition"
          )}
        </button>
      </div>
    </form>
  );
};

export default function ArtDetailPage() {
  const router = useRouter();
  const params = useParams();
  const artId = params?.id as string;
  const cropperRef = useRef<CropperRef>(null);

  const [art, setArt] = useState<Art | null>(null);
  const [compositions, setCompositions] = useState<Composition[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCompositionForm, setShowCompositionForm] = useState(true);

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

  // Add new state for composition modal
  const [showCompositionModal, setShowCompositionModal] = useState(false);

  // Fetch art, compositions and check for pending status
  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const currentUser = await getCurrentUser();
        if (!artId) return;

        const artResourceName = `${currentUser.name}/arts/${artId}`;

        // First get the art to ensure it exists
        const artData = await getArt(artResourceName);
        setArt(artData);

        // Only fetch compositions if we have a valid art
        if (artData) {
          const compositionsResponse = await listCompositions({
            parent: artResourceName,
            pageSize: 100, // Add page size as it's required by the proto
          });
          setCompositions(compositionsResponse.compositions || []);

          // Hide form if there's a pending composition
          const hasPendingComposition = compositionsResponse.compositions?.some(
            (comp) => comp.status === 1 || comp.status === 2 // PENDING or PROCESSING
          );
          setShowCompositionForm(!hasPendingComposition);
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

  // Poll for pending compositions
  useEffect(() => {
    if (!params?.id || !compositions || !art) return;

    const hasPendingCompositions = compositions.some(
      (comp) => comp.status === 1 || comp.status === 2
    );

    if (hasPendingCompositions) {
      const interval = setInterval(async () => {
        try {
          // Use the actual art name from the art object instead of constructing it manually
          const response = await listCompositions({
            parent: art.name,
            pageSize: 100,
          });
          setCompositions(response.compositions);

          // Check if there are still pending compositions
          const stillPending = response.compositions.some(
            (comp) => comp.status === 1 || comp.status === 2
          );

          // If no more pending compositions, clear the interval
          if (!stillPending) {
            clearInterval(interval);
          }
        } catch (error) {
          console.error("Error polling compositions:", error);
        }
      }, 2000); // Poll every 2 seconds

      return () => clearInterval(interval);
    }
  }, [params?.id, compositions, art]);

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
          "x-amz-acl": "private",
          "x-forwarded-proto": "http",
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

  const CompositionCard = ({ composition }: { composition: Composition }) => {
    const [isExpanded, setIsExpanded] = useState(false);
    const getStatusInfo = (status: number) => {
      switch (status) {
        case 1:
          return { text: "Pending", color: "text-yellow-400" };
        case 2:
          return { text: "Processing", color: "text-blue-400" };
        case 3:
          return { text: "Complete", color: "text-green-400" };
        case 4:
          return { text: "Failed", color: "text-red-400" };
        default:
          return { text: "Unknown", color: "text-gray-400" };
      }
    };

    const statusInfo = getStatusInfo(composition.status);

    return (
      <div className="bg-dark-300 rounded-lg p-4 mb-4">
        <div className="flex justify-between items-center mb-4">
          <div className={`${statusInfo.color} font-medium`}>
            {statusInfo.text}
          </div>
          <div className="text-sm text-slate-400">
            {new Date(
              Number(composition.createTime?.seconds) * 1000
            ).toLocaleString()}
          </div>
        </div>

        {composition.previewUrl ? (
          <div className="aspect-square w-full mb-4">
            <Image
              src={composition.previewUrl}
              alt="Composition preview"
              width={400}
              height={400}
              className="w-full h-full object-contain rounded-full"
            />
          </div>
        ) : (
          <div className="aspect-square w-full mb-4 bg-dark-400 rounded-full flex items-center justify-center">
            {composition.status === 1 || composition.status === 2 ? (
              <div className="animate-pulse text-slate-500">Processing...</div>
            ) : composition.status === 3 ? (
              <div className="text-orange-400 text-center p-4">
                <p className="font-medium mb-2">Image Generation Failed</p>
                <p className="text-sm">
                  The preview image couldn&apos;t be generated.
                </p>
              </div>
            ) : composition.status === 4 ? (
              <div className="text-red-400 text-center p-4">
                <p className="font-medium mb-2">Generation Failed</p>
                <p className="text-sm">
                  {composition.errorMessage || "Unknown error occurred"}
                </p>
              </div>
            ) : (
              <div className="text-slate-400 text-center p-4">
                <p className="font-medium">No Preview Available</p>
              </div>
            )}
          </div>
        )}

        {composition.status === 3 && (
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-slate-400">Thread Length:</span>
              <span className="text-slate-300">
                {composition.threadLength}m
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-slate-400">Total Lines:</span>
              <span className="text-slate-300">{composition.totalLines}</span>
            </div>
            <div className="flex space-x-2 mt-4">
              {composition.gcodeUrl && (
                <a
                  href={composition.gcodeUrl}
                  download={`thread-art-${composition.name
                    .split("/")
                    .pop()}.gcode`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex-1 px-3 py-2 bg-primary-500 text-white text-center rounded hover:bg-primary-600 transition-colors text-sm"
                >
                  Download GCode
                </a>
              )}
              {composition.pathlistUrl && (
                <a
                  href={composition.pathlistUrl}
                  download={`thread-art-${composition.name
                    .split("/")
                    .pop()}.txt`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex-1 px-3 py-2 bg-primary-500 text-white text-center rounded hover:bg-primary-600 transition-colors text-sm"
                >
                  Download Paths
                </a>
              )}
            </div>
          </div>
        )}

        {/* Settings Accordion */}
        <div className="mt-4 border-t border-dark-400 pt-4">
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="w-full flex justify-between items-center text-slate-300 hover:text-slate-200"
          >
            <span className="text-sm font-medium">Settings</span>
            <svg
              className={`w-5 h-5 transform transition-transform ${
                isExpanded ? "rotate-180" : ""
              }`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 9l-7 7-7-7"
              />
            </svg>
          </button>
          {isExpanded && (
            <div className="mt-4 space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-slate-400">Nails Quantity:</span>
                <span className="text-slate-300">
                  {composition.nailsQuantity}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Image Size:</span>
                <span className="text-slate-300">{composition.imgSize}px</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Max Paths:</span>
                <span className="text-slate-300">{composition.maxPaths}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Starting Nail:</span>
                <span className="text-slate-300">
                  {composition.startingNail}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Minimum Difference:</span>
                <span className="text-slate-300">
                  {composition.minimumDifference}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Brightness Factor:</span>
                <span className="text-slate-300">
                  {composition.brightnessFactor}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Image Contrast:</span>
                <span className="text-slate-300">
                  {composition.imageContrast}%
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Physical Radius:</span>
                <span className="text-slate-300">
                  {composition.physicalRadius}mm
                </span>
              </div>
            </div>
          )}
        </div>
      </div>
    );
  };

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-12">
        <div className="max-w-6xl mx-auto">
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
        <div className="max-w-6xl mx-auto">
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
  const needsImage = art.status === 1;
  const created = art.createTime
    ? new Date(Number(art.createTime.seconds) * 1000)
    : new Date();

  return (
    <div className="container mx-auto px-4 py-12">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold text-slate-100">{art.title}</h1>
            <p className="text-slate-400 text-sm mt-1">
              Created on {created.toLocaleDateString()} at{" "}
              {created.toLocaleTimeString()}
            </p>
          </div>
          <div className="flex items-center space-x-4">
            <span
              className={`${statusInfo.color} px-3 py-1 rounded-full text-sm border border-current`}
            >
              {statusInfo.text}
            </span>
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

        {/* Main content */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Left column - Art image */}
          <div>
            {needsImage ? (
              <div className="bg-dark-200 rounded-lg p-6">
                <ErrorMessage message={uploadError} />
                <SuccessMessage
                  message={
                    uploadSuccess ? "Image uploaded successfully!" : null
                  }
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
                    className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer ${
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
                    </div>
                  </div>
                )}
              </div>
            ) : art.imageUrl ? (
              <div className="bg-dark-200 rounded-lg p-6">
                <div className="rounded-lg overflow-hidden bg-dark-300 aspect-square">
                  <Image
                    src={art.imageUrl}
                    alt={art.title}
                    className="w-full h-full object-contain"
                    priority={true}
                    width={1000}
                    height={1000}
                    onError={(e) => {
                      console.error("Image failed to load:", art.imageUrl);
                      e.currentTarget.src =
                        "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100' height='100' viewBox='0 0 100 100'%3E%3Crect width='100' height='100' fill='%23444'/%3E%3Ctext x='50' y='50' font-family='Arial' font-size='12' text-anchor='middle' fill='%23fff'%3EImage Error%3C/text%3E%3C/svg%3E";
                    }}
                  />
                </div>
              </div>
            ) : (
              <div className="bg-dark-200 rounded-lg p-6">
                <div className="rounded-lg bg-dark-300 aspect-square flex items-center justify-center">
                  <p className="text-slate-300">No image available</p>
                </div>
              </div>
            )}
          </div>

          {/* Right column - Compositions */}
          <div>
            {art.imageUrl && (
              <div className="bg-dark-200 rounded-lg p-6">
                <div className="flex justify-between items-center mb-6">
                  <h2 className="text-xl font-semibold text-slate-100">
                    Compositions
                  </h2>
                  {showCompositionForm && (
                    <button
                      onClick={() => setShowCompositionModal(true)}
                      className="px-4 py-2 bg-primary-500 text-white rounded hover:bg-primary-600 transition-colors"
                    >
                      New Composition
                    </button>
                  )}
                </div>

                {compositions.length > 0 ? (
                  <div className="space-y-4 max-h-[800px] overflow-y-auto pr-2">
                    {[...compositions]
                      .sort((a, b) => {
                        const aTime = Number(a.createTime?.seconds) || 0;
                        const bTime = Number(b.createTime?.seconds) || 0;
                        return bTime - aTime;
                      })
                      .map((composition) => (
                        <CompositionCard
                          key={composition.name}
                          composition={composition}
                        />
                      ))}
                  </div>
                ) : (
                  <p className="text-slate-400 text-center py-8">
                    No compositions yet. Create your first one!
                  </p>
                )}
              </div>
            )}
          </div>
        </div>

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

        {/* New composition modal */}
        {showCompositionModal && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-dark-200 rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
              <div className="flex justify-between items-center mb-6">
                <h3 className="text-xl font-semibold text-slate-100">
                  Create New Composition
                </h3>
                <button
                  onClick={() => setShowCompositionModal(false)}
                  className="text-slate-400 hover:text-slate-300"
                >
                  <svg
                    className="w-6 h-6"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              </div>

              <CompositionForm
                onSubmit={async (formData) => {
                  try {
                    await createComposition({
                      parent: art.name,
                      composition: {
                        nailsQuantity: formData.nailsQuantity,
                        imgSize: formData.imgSize,
                        maxPaths: formData.maxPaths,
                        startingNail: formData.startingNail,
                        minimumDifference: formData.minimumDifference,
                        brightnessFactor: formData.brightnessFactor,
                        imageContrast: formData.imageContrast,
                        physicalRadius: formData.physicalRadius,
                      },
                    });

                    // Fetch latest compositions immediately
                    const compositionsResponse = await listCompositions({
                      parent: art.name,
                      pageSize: 100,
                    });
                    setCompositions(compositionsResponse.compositions || []);
                    setShowCompositionForm(false);
                    setShowCompositionModal(false);
                  } catch (error) {
                    console.error("Failed to create composition:", error);
                    setError("Failed to create composition. Please try again.");
                  }
                }}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
