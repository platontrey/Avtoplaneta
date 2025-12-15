    // Web Worker for image cropping
self.onmessage = function(e) {
  console.log('Worker received message:', e.data);
  const { imageSrc, crop, zoom, position, naturalWidth, naturalHeight, displayWidth, displayHeight } = e.data;

  if (!self.OffscreenCanvas) {
    self.postMessage({ success: false, error: 'OffscreenCanvas not supported' });
    return;
  }

  // Fetch the image and create bitmap
  fetch(imageSrc)
    .then(response => {
      if (!response.ok) {
        throw new Error('Failed to fetch image');
      }
      return response.blob();
    })
    .then(blob => createImageBitmap(blob))
    .then(bitmap => {
      console.log('Bitmap created in worker, dimensions:', bitmap.width, bitmap.height);
      try {
        // Create OffscreenCanvas
        const canvas = new OffscreenCanvas(crop.width / zoom, crop.height / zoom);
        console.log('Canvas created, size:', canvas.width, canvas.height);
        const ctx = canvas.getContext('2d');
        if (!ctx) {
          throw new Error('No 2d context');
        }

        // Account for zoom and position in crop coordinates
        const adjustedCrop = {
          x: (crop.x - position.x) / zoom,
          y: (crop.y - position.y) / zoom,
          width: crop.width / zoom,
          height: crop.height / zoom,
        };

        const scaleX = naturalWidth / displayWidth;
        const scaleY = naturalHeight / displayHeight;
        console.log('Adjusted crop:', adjustedCrop, 'scales:', scaleX, scaleY);

        // Draw the cropped image
        ctx.drawImage(
          bitmap,
          adjustedCrop.x * scaleX,
          adjustedCrop.y * scaleY,
          adjustedCrop.width * scaleX,
          adjustedCrop.height * scaleY,
          0,
          0,
          adjustedCrop.width,
          adjustedCrop.height
        );
        console.log('drawImage completed');

        // Convert to blob
        canvas.convertToBlob({ type: 'image/jpeg', quality: 0.95 }).then(blob => {
          console.log('Blob created, size:', blob.size);
          self.postMessage({ success: true, blob });
        }).catch(error => {
          console.error('convertToBlob error:', error);
          self.postMessage({ success: false, error: error.message });
        });
      } catch (error) {
        console.error('Worker processing error:', error);
        self.postMessage({ success: false, error: error.message });
      }
    })
    .catch(error => {
      console.error('Fetch or bitmap error:', error);
      self.postMessage({ success: false, error: error.message });
    });
};