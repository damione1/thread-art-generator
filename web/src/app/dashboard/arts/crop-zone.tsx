import React, { useState } from 'react'
import Cropper from 'react-easy-crop'
import Slider from '@mui/material/Slider';
import getCroppedImg from './crop-zone-functions';

export function CropZone({
  showCropZone,
  setShowCropZone,
  originalImageBlob,
  setOriginalImageBlob,
  croppedImageBlob,
  setCroppedImageBlob
}: {
  showCropZone: boolean,
  setShowCropZone: (showCropZone: boolean) => void,
  originalImageBlob: Blob | null,
  setOriginalImageBlob: (imageBlob: Blob) => void,
  croppedImageBlob: Blob | null,
  setCroppedImageBlob: (imageBlob: Blob) => void}
) {


const [crop, setCrop] = useState({ x: 0, y: 0 })
const [rotation, setRotation] = useState(0)
const [zoom, setZoom] = useState(1)
const [croppedAreaPixels, setCroppedAreaPixels] = useState(null)
const [croppedImage, setCroppedImage] = useState(null)

const  handleCropComplete = async () => {
  console.log("Crop complete")
  try {
    if (!originalImageBlob || !croppedAreaPixels) {
      return
    }
    const croppedImageBlob = await getCroppedImg(
      URL.createObjectURL(originalImageBlob),
      croppedAreaPixels,
      rotation
    )
    if (!croppedImageBlob) {
      return
    }
    setCroppedImageBlob(croppedImageBlob)
    setShowCropZone(false)
  } catch (e) {
    console.error(e)
  }
}


const onCropComplete = (croppedArea, croppedAreaPixels) => {
  setCroppedAreaPixels(croppedAreaPixels)
}

const onCropChange = (crop) => {
  setCrop(crop)
}

const onZoomChange = (zoom) => {
  setZoom(zoom)
}


const onClose = () => {
  setCroppedImage(null)
}


return (
  showCropZone && originalImageBlob && (
    <>
      <div className="crop-container">
        <Cropper
          image={URL.createObjectURL(originalImageBlob)}
          crop={crop}
          zoom={zoom}
          aspect={1}
          cropShape="round"
          showGrid={false}
          onCropChange={onCropChange}
          onCropComplete={onCropComplete}
          onZoomChange={onZoomChange}
        />
    </div>
    <div className="controls">
        <Slider
            aria-label="Zoom"
            value={zoom}
            min={1}
            max={3}
            step={0.1}
            color="secondary"
            onChange={(e, zoom) => setZoom(zoom as number)}
        />
        <button onClick={handleCropComplete}>Complete Crop</button>
    </div>
  </>
  )
)
}
