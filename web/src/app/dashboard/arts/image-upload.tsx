import { useSession } from 'next-auth/react';
import React, { useState, useCallback, useEffect } from 'react';
import Dropzone, { useDropzone } from 'react-dropzone';
import Cropper, { Area, Point } from 'react-easy-crop';
import Slider from '@mui/material/Slider';
import Image from 'next/image';
import { DropZone } from './drop-zone';
import { CropIcon, EditIcon, Trash2, UploadIcon } from 'lucide-react';
import { CropZone } from './crop-zone';
import { Button } from '@mui/material';

const ImageCropUpload = ({ sourceImage, handleImageChange }: { sourceImage: string, handleImageChange: (imageBlob: Blob) =>  Promise<boolean> })  => {
    const [originalImageBlob, setOriginalImageBlob] = useState<Blob | null>(null);
    const [croppedImageBlob, setCroppedImageBlob] = useState<Blob | null>(null);
    const [showCropZone, setShowCropZone] = useState(false);
    const [isImageUploaded, setIsImageUploaded] = useState(false);

    // When a source image is provided, we need to convert it to a blob
    useEffect(() => {
        if (sourceImage) {
            fetch(sourceImage)
                .then(res => res.blob())
                .then(blob => setCroppedImageBlob(blob))
                .then((blob) => setIsImageUploaded(true))
        }
    } , [sourceImage]);

    const handleDrop =  function(acceptedFiles: File[]) {
        const imageBlob = new Blob([acceptedFiles[0]]);
        setOriginalImageBlob(imageBlob);
        setShowCropZone(true);
        setIsImageUploaded(false);
    };

    const handleDelete = function() {
        setOriginalImageBlob(null);
        setCroppedImageBlob(null);
    }

    const handleSubmit = function() {
        if (croppedImageBlob) {
            handleImageChange(croppedImageBlob).then((success) => {
                if (success) {
                    setIsImageUploaded(true);
                }
            });

            setShowCropZone(false);
        }
    }


    return (
        <div>
            {(croppedImageBlob) && (
                //If we have an original image blob, we can display it
                <div>
                    <Image src={URL.createObjectURL(croppedImageBlob)} width={400} height={400}  alt="image" style={{ borderRadius: '50%' }} />
                    <button className="inline-flex items-center justify-center rounded-md bg-primary px-2 py-4 text-center mx-1 mt-5 font-medium text-white hover:bg-opacity-90" onClick={handleDelete}><Trash2   /></button>
                    {originalImageBlob  && <button className="inline-flex items-center justify-center rounded-md bg-primary px-2 py-4 text-center mx-1 mt-5 font-medium text-white hover:bg-opacity-90" onClick={() => setShowCropZone(true)}> <CropIcon /> </button>}
                    {!isImageUploaded  && <button className="inline-flex items-center justify-center rounded-md bg-primary px-2 py-4 text-center mx-1 mt-5 font-medium text-white hover:bg-opacity-90 lg:px-8 xl:px-10" onClick={() => handleSubmit()}><UploadIcon /> Upload image</button>}
                </div>
            )}
            {!croppedImageBlob && (
                <DropZone handleDrop={handleDrop} />
            )}
            <CropZone setShowCropZone={setShowCropZone} showCropZone={showCropZone} originalImageBlob={originalImageBlob} setOriginalImageBlob={setOriginalImageBlob} croppedImageBlob={croppedImageBlob} setCroppedImageBlob={setCroppedImageBlob} />
        </div>
    );
};

export default ImageCropUpload;
