import React, {useCallback, useMemo} from 'react'
import {useDropzone} from 'react-dropzone'

const baseStyle = {
    flex: 1,
    alignItems: 'center',
    padding: '20px',
    borderWidth: 2,
    borderRadius: 2,
    borderColor: '#eeeeee',
    borderStyle: 'dashed',
    backgroundColor: null,
    color: '#bdbdbd',
    outline: 'none',
    transition: 'border .24s ease-in-out'
  };

  const focusedStyle = {
    borderColor: '#2196f3'
  };

  const acceptStyle = {
    borderColor: '#00e676'
  };

  const rejectStyle = {
    borderColor: '#ff1744'
  };


export function DropZone(
    {
    handleDrop
}: {
    handleDrop: (acceptedFiles: File[]) => void
}) {

const {
    getRootProps,
    getInputProps,
    acceptedFiles,
    isFocused,
    isDragAccept,
    isDragReject
  } = useDropzone({
    accept: {'image/*': []},
    maxFiles:1,
    onDrop: handleDrop,
});


  const style = useMemo(() => ({
    ...baseStyle,
    ...(isFocused ? focusedStyle : {}),
    ...(isDragAccept ? acceptStyle : {}),
    ...(isDragReject ? rejectStyle : {})
  }), [
    isFocused,
    isDragAccept,
    isDragReject
  ]);

return (
    <div className="container">
        <div {...getRootProps({style: {...style, backgroundColor: undefined}})}>
            <input {...getInputProps()} />
            <p>Drag 'n' drop the image here, or click to select the image</p>
        </div>
    </div>
)
}
